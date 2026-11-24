package jobs

import (
	"crypto/rand"
	"encoding/hex"
)

const (
	idBytes      = 16
	versionByte  = 6
	variantByte  = 8
	versionMask  = 0x0f
	versionFour  = 0x40
	variantMask  = 0x3f
	variantRFC   = 0x80
	shadowBucket = 4
)

// newID is a random UUID version 4 in hex, without dashes.
func newID() string {
	var id [idBytes]byte
	_, _ = rand.Read(id[:])
	id[versionByte] = id[versionByte]&versionMask | versionFour
	id[variantByte] = id[variantByte]&variantMask | variantRFC

	return hex.EncodeToString(id[:])
}
