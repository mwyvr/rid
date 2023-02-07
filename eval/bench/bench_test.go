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
var resultRID rid.ID

func BenchmarkRid(b *testing.B) {
	var r rid.ID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = rid.New()
		}
		resultRID = r
	})
}

// https://github.com/rs/xid
// xid uses a random-initialized (once only) monotonically increasing counter
var resultXID xid.ID

func BenchmarkXid(b *testing.B) {
	var r xid.ID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = xid.New()
		}
		resultXID = r
	})
}

// https://github.com/segmentio/ksuid
// ksuid ids incorporate crypto/rand generated numbers

var resultKSUID ksuid.KSUID

func BenchmarkKsuid(b *testing.B) {
	var r ksuid.KSUID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = ksuid.New()
		}
		resultKSUID = r
	})
}

// uuid ids incorporate crypto/rand generated numbers
var resultUUID guuid.UUID

func BenchmarkGoogleUuid(b *testing.B) {
	var r guuid.UUID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// https://pkg.go.dev/github.com/google/UUID#NewRandom
			// uuid v4, equiv to NewRandom()
			r = guuid.New()
		}
		resultUUID = r
	})
}

// as configured here, for a good comparison, ulid ids incorporate crypto/rand
// generated numbers
var resultULID ulid.ULID

func BenchmarkUlid(b *testing.B) {
	var r ulid.ULID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t := time.Now().UTC()
			entropy := rand.New(rand.NewSource(t.UnixNano()))
			r = ulid.MustNew(ulid.Timestamp(t), entropy)
		}
		resultULID = r
	})
}

// https://github.com/kjk/betterguid
// like rs/xid, uses a monotonically incrementing counter rather than
// true randomness
var resultBGUID string

func BenchmarkBetterguid(b *testing.B) {
	var r string
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = betterguid.New()
		}
		resultBGUID = r
	})
}
