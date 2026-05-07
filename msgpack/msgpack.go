package msgpack

import (
	"bytes"
	"io"

	mp "github.com/vmihailenco/msgpack/v5"
)

// Marshal encodes v into MessagePack binary format.
func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := mp.NewEncoder(&buf)
	enc.SetSortMapKeys(true)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal decodes MessagePack data into v.
func Unmarshal(data []byte, v interface{}) error {
	dec := mp.NewDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}

// Encoder writes MessagePack-encoded values to an output stream.
type Encoder struct {
	enc *mp.Encoder
}

// NewEncoder returns an Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	enc := mp.NewEncoder(w)
	enc.SetSortMapKeys(true)
	return &Encoder{enc: enc}
}

// Encode writes the MessagePack encoding of v to the stream.
func (e *Encoder) Encode(v interface{}) error {
	return e.enc.Encode(v)
}

// Decoder reads MessagePack-encoded values from an input stream.
type Decoder struct {
	dec *mp.Decoder
}

// NewDecoder returns a Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{dec: mp.NewDecoder(r)}
}

// Decode reads the next MessagePack value from the stream and stores it in v.
func (d *Decoder) Decode(v interface{}) error {
	return d.dec.Decode(v)
}
