package bench

import (
	"math/rand"
	"testing"
	"time"

	guuid "github.com/google/uuid"
	"github.com/kjk/betterguid"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
	"github.com/solutionroute/rid"
)

// rid ids incorporate crypto/rand generated numbers
func BenchmarkRid(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rid.New()
		}
	})
}

// https://github.com/rs/xid
// xid uses a random-initialized (once only) monotonically increasing counter
func BenchmarkXid(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = xid.New()
		}
	})
}

// https://github.com/segmentio/ksuid
// ksuid ids incorporate crypto/rand generated numbers
func BenchmarkKsuid(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = ksuid.New()
		}
	})
}

// uuid ids incorporate crypto/rand generated numbers
func BenchmarkGoogleUuid(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// https://pkg.go.dev/github.com/google/UUID#NewRandom
			// uuid v4, equiv to NewRandom()
			_ = guuid.New()
		}
	})
}

// as configured here, for a good comparison, ulid ids incorporate crypto/rand
// generated numbers
func BenchmarkUlid(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t := time.Now().UTC()
			entropy := rand.New(rand.NewSource(t.UnixNano()))
			_ = ulid.MustNew(ulid.Timestamp(t), entropy)
		}
	})
}

// https://github.com/kjk/betterguid
// like rs/xid, uses a monotonically incrementing counter rather than
// true randomness
func BenchmarkBetterguid(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = betterguid.New()
		}
	})
}
