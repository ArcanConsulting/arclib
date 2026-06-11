package crypto

import (
	"lukechampine.com/adiantum"
	"lukechampine.com/adiantum/hbsh"
)

// SQLiteVFSKeySize is the key length, in bytes, that SQLiteVFSCipher expects.
const SQLiteVFSKeySize = 32

// SQLiteVFSCipher builds the length-preserving HBSH cipher used for at-rest
// SQLite page encryption: Adiantum over XChaCha12 (the wide-block construction
// Android uses for storage encryption), from a 32-byte key. It returns nil for
// any other key length.
//
// This centralises the at-rest cipher choice in arclib ("all crypto via
// arclib"): a consumer plugs the returned cipher into a SQLite encrypting VFS
// (e.g. the ncruces/go-sqlite3 adiantum VFS via its HBSHCreator hook) and owns
// only the VFS plumbing, not the cipher. The construction is identical to that
// VFS's stock default, so databases encrypted with the stock VFS open
// unchanged — there is no on-disk format change.
func SQLiteVFSCipher(key []byte) *hbsh.HBSH {
	if len(key) != SQLiteVFSKeySize {
		return nil
	}
	return adiantum.New(key)
}
