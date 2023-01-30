package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/kjk/betterguid"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
	"github.com/solutionroute/rid"
)

type pkg struct {
	name       string
	blen       int
	elen       int
	ksortable  bool
	sample     string
	next       string
	uniq       string
	components string
}

func main() {
	packages := []pkg{
		{
			"[solutionroute/rid](https://github.com/solutionroute/rid)",
			len(rid.New().Bytes()),
			len(rid.New().String()),
			true,
			rid.New().String(),
			rid.New().String(),
			"fastrand",
			"4 byte ts(sec) : 6 byte random",
		},
		{
			"[rs/xid](https://github.com/rs/xid)",
			len(xid.New().Bytes()),
			len(xid.New().String()),
			true,
			xid.New().String(),
			xid.New().String(),
			"counter",
			"4 byte ts(sec) : 2 byte mach ID : 2 byte pid : 3 byte monotonic counter",
		},
		{
			"[segmentio/ksuid](https://github.com/segmentio/ksuid)",
			len(ksuid.New().Bytes()),
			len(ksuid.New().String()),
			true,
			ksuid.New().String(),
			ksuid.New().String(),
			"random",
			"4 byte ts(sec) : 16 byte random",
		},
		{
			"[google/uuid](https://github.com/google/uuid)",
			len(uuid.New()),
			len(uuid.New().String()),
			false,
			uuid.New().String(),
			uuid.New().String(),
			"crypt/rand",
			"v4: 16 bytes random with version & variant embedded",
		},
		{
			"[oklog/ulid](https://github.com/oklog/ulid)",
			len(newUlid()),
			len(newUlid().String()),
			true,
			newUlid().String(),
			newUlid().String(),
			"crypt/rand",
			"6 byte ts(ms) : 10 byte counter random init per ts(ms)",
		},
		{
			"[kjk/betterguid](https://github.com/kjk/betterguid)",
			8 + 9, // only available as a string
			len(betterguid.New()),
			true,
			betterguid.New(),
			betterguid.New(),
			"counter",
			"8 byte ts(ms) : 9 byte counter random init per ts(ms)",
		},
	}

	fmt.Printf("| Package                                                   |BLen|ELen| K-Sort| Encoded ID and Next | Method | Components |\n")
	fmt.Printf("|-----------------------------------------------------------|----|----|-------|---------------------|--------|------------|\n")
	for _, v := range packages {
		fmt.Printf("| %-57s | %d | %d | %5v | `%s`<br>`%s` | %s | %s |\n",
			v.name, v.blen, v.elen, v.ksortable, v.sample, v.next, v.uniq, v.components)
	}
}

// ulid is configured here to be similar (random component) to rid, ksuid, uuid
func newUlid() ulid.ULID {
	t := time.Now().UTC()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	return ulid.MustNew(ulid.Timestamp(t), entropy)
}
