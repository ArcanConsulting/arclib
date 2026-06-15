package transport

import (
	"errors"
	"fmt"

	"arcan-it.de/arclib/crypto"
	"arcan-it.de/arclib/msgpack"
	"arcan-it.de/arclib/protocol"
)

// Handshake errors.
var (
	// ErrHandshakeUnexpectedOp is returned when a handshake frame carries an
	// opcode other than the expected SESSION_INIT / SESSION_ACK.
	ErrHandshakeUnexpectedOp = errors.New("transport: unexpected handshake opcode")
	// ErrHandshakeBadKey is returned when handshake key material has the wrong size.
	ErrHandshakeBadKey = errors.New("transport: malformed handshake key material")
)

// kexInfo binds the derived secret to this transport's hybrid KEX construction.
const kexInfo = "myclerk/transport hybrid-kex v1"

// handshakeInit is the SESSION_INIT (0x0003) payload sent by the initiator.
type handshakeInit struct {
	X25519Pub  []byte `msgpack:"x"`  // 32-byte ephemeral X25519 public key
	MLKEMEncap []byte `msgpack:"mk"` // 1184-byte ML-KEM-768 encapsulation key
	ClientSalt []byte `msgpack:"cs"` // 16-byte session salt
}

// handshakeAck is the SESSION_ACK (0x0004) payload sent by the responder.
type handshakeAck struct {
	X25519Pub       []byte `msgpack:"x"`  // 32-byte ephemeral X25519 public key
	MLKEMCiphertext []byte `msgpack:"mc"` // 1088-byte ML-KEM-768 ciphertext
	ServerSalt      []byte `msgpack:"ss"` // 16-byte session salt
}

type writeFrameFunc func(protocol.OpCode, []byte) error
type readFrameFunc func() (protocol.OpCode, []byte, error)

// clientHandshake runs the SESSION_INIT/SESSION_ACK initiator side (hybrid
// X25519 + ML-KEM-768) and returns the established Tier-3 session.
func clientHandshake(write writeFrameFunc, read readFrameFunc) (*protocol.AeadSession, error) {
	cx, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		return nil, fmt.Errorf("transport: x25519 keygen: %w", err)
	}
	cm, err := crypto.GenerateMLKEMKeyPair()
	if err != nil {
		return nil, fmt.Errorf("transport: ml-kem keygen: %w", err)
	}
	clientSalt, err := protocol.GenerateSessionSalt()
	if err != nil {
		return nil, err
	}

	payload, err := msgpack.Marshal(handshakeInit{
		X25519Pub:  cx.Public[:],
		MLKEMEncap: cm.EncapsulationKeyBytes(),
		ClientSalt: clientSalt,
	})
	if err != nil {
		return nil, fmt.Errorf("transport: encode init: %w", err)
	}
	if err := write(protocol.OpSessionInit, payload); err != nil {
		return nil, err
	}

	op, ackPayload, err := read()
	if err != nil {
		return nil, err
	}
	if op != protocol.OpSessionAck {
		return nil, ErrHandshakeUnexpectedOp
	}
	var ack handshakeAck
	if err := msgpack.Unmarshal(ackPayload, &ack); err != nil {
		return nil, fmt.Errorf("transport: decode ack: %w", err)
	}
	serverPub, err := toX25519(ack.X25519Pub)
	if err != nil {
		return nil, err
	}
	if len(ack.ServerSalt) != protocol.SessionSaltSize {
		return nil, ErrHandshakeBadKey
	}

	xss, err := cx.ECDH(serverPub)
	if err != nil {
		return nil, fmt.Errorf("transport: x25519 ecdh: %w", err)
	}
	mss, err := cm.Decapsulate(ack.MLKEMCiphertext)
	if err != nil {
		return nil, fmt.Errorf("transport: ml-kem decapsulate: %w", err)
	}
	secret, err := deriveSharedSecret(xss, mss, clientSalt, ack.ServerSalt)
	if err != nil {
		return nil, err
	}
	return protocol.NewAeadSession(secret, clientSalt, ack.ServerSalt, true)
}

// serverHandshake runs the responder side: it reads SESSION_INIT, replies with
// SESSION_ACK, and returns the established Tier-3 session.
func serverHandshake(write writeFrameFunc, read readFrameFunc) (*protocol.AeadSession, error) {
	op, initPayload, err := read()
	if err != nil {
		return nil, err
	}
	if op != protocol.OpSessionInit {
		return nil, ErrHandshakeUnexpectedOp
	}
	var init handshakeInit
	if err := msgpack.Unmarshal(initPayload, &init); err != nil {
		return nil, fmt.Errorf("transport: decode init: %w", err)
	}
	clientPub, err := toX25519(init.X25519Pub)
	if err != nil {
		return nil, err
	}
	if len(init.ClientSalt) != protocol.SessionSaltSize {
		return nil, ErrHandshakeBadKey
	}

	sx, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		return nil, fmt.Errorf("transport: x25519 keygen: %w", err)
	}
	serverSalt, err := protocol.GenerateSessionSalt()
	if err != nil {
		return nil, err
	}
	mss, ct, err := crypto.MLKEMEncapsulate(init.MLKEMEncap)
	if err != nil {
		return nil, fmt.Errorf("transport: ml-kem encapsulate: %w", err)
	}
	xss, err := sx.ECDH(clientPub)
	if err != nil {
		return nil, fmt.Errorf("transport: x25519 ecdh: %w", err)
	}
	secret, err := deriveSharedSecret(xss, mss, init.ClientSalt, serverSalt)
	if err != nil {
		return nil, err
	}

	payload, err := msgpack.Marshal(handshakeAck{
		X25519Pub:       sx.Public[:],
		MLKEMCiphertext: ct,
		ServerSalt:      serverSalt,
	})
	if err != nil {
		return nil, fmt.Errorf("transport: encode ack: %w", err)
	}
	if err := write(protocol.OpSessionAck, payload); err != nil {
		return nil, err
	}
	return protocol.NewAeadSession(secret, init.ClientSalt, serverSalt, false)
}

// deriveSharedSecret combines the classical and post-quantum shared secrets into
// a single 32-byte key via HKDF, salted by both endpoints' handshake salts.
func deriveSharedSecret(x25519 [crypto.SharedSecretSize]byte, mlkem [crypto.MLKEMSharedKeySize]byte, clientSalt, serverSalt []byte) ([]byte, error) {
	ikm := make([]byte, 0, len(x25519)+len(mlkem))
	ikm = append(ikm, x25519[:]...)
	ikm = append(ikm, mlkem[:]...)
	salt := make([]byte, 0, len(clientSalt)+len(serverSalt))
	salt = append(salt, clientSalt...)
	salt = append(salt, serverSalt...)
	return crypto.HKDFDeriveKey(ikm, salt, []byte(kexInfo), crypto.KeySize)
}

func toX25519(b []byte) ([crypto.X25519KeySize]byte, error) {
	var k [crypto.X25519KeySize]byte
	if len(b) != crypto.X25519KeySize {
		return k, ErrHandshakeBadKey
	}
	copy(k[:], b)
	return k, nil
}
