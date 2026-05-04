package crypto

import (
	"crypto/sha256"
	"hash/crc32"
)

const (
	// SHA256Size is the size of a SHA-256 hash in bytes.
	SHA256Size = sha256.Size // 32 bytes

	// CRC16Size is the size of a CRC-16 checksum in bytes.
	CRC16Size = 2

	// CRC32Size is the size of a CRC-32 checksum in bytes.
	CRC32Size = 4

	// crc16Polynomial is the CRC-16-IBM (reflected) polynomial.
	// Also known as CRC-16-ANSI or CRC-16-Modbus.
	crc16Polynomial = 0xA001

	// crc16Init is the initial value for CRC-16-IBM computation.
	crc16Init = 0xFFFF
)

// SHA256Sum returns the SHA-256 hash of the given data.
func SHA256Sum(data []byte) [SHA256Size]byte {
	return sha256.Sum256(data)
}

// CRC16 computes the CRC-16-IBM checksum of the given data.
//
// Uses polynomial 0xA001 (reflected form of 0x8005) with initial value 0xFFFF.
// This is the variant used in the MyClerk Protocol for Tier 1 integrity checks.
func CRC16(data []byte) uint16 {
	crc := uint16(crc16Init)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ crc16Polynomial
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// CRC32 computes the IEEE CRC-32 checksum of the given data.
//
// Uses the standard IEEE polynomial (0x04C11DB7), compatible with
// Ethernet, PKZIP, and most other CRC-32 implementations.
func CRC32(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}
