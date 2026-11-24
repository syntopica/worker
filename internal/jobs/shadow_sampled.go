package jobs

import (
	"crypto/sha256"
	"encoding/binary"
)

const shadowBuckets = 1 << 32

func shadowSampled(id string, rate float64) bool {
	sum := sha256.Sum256([]byte(id))
	return float64(binary.BigEndian.Uint32(sum[:])) < rate*shadowBuckets
}
