// Package main produces for documentation a markdown formatted table
// illustrating key differences between a number of unique ID packages.
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/kjk/betterguid"
	"github.com/mwyvr/rid"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
)

type pkg struct {
	name       string
	blen       int
	elen       int
	ksortable  bool
	sample     string
	next       string
	next2      string
	next3      string
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
			rid.New().String(),
			rid.New().String(),
			"crypto/rand",
			"4 byte ts(sec) : 6 byte random",
		},
		{
			"[rs/xid](https://github.com/rs/xid)",
			len(xid.New().Bytes()),
			len(xid.New().String()),
			true,
			xid.New().String(),
			xid.New().String(),
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
			ksuid.New().String(),
			ksuid.New().String(),
			"math/rand",
			"4 byte ts(sec) : 16 byte random",
		},
		{
			"[google/uuid](https://github.com/google/uuid)",
			len(uuid.New()),
			len(uuid.New().String()),
			false,
			uuid.New().String(),
			uuid.New().String(),
			uuid.New().String(),
			uuid.New().String(),
			"crypt/rand",
			"v4: 16 bytes random with version & variant embedded",
		},
		{
			"[google/uuid](https://github.com/google/uuid)",
			len(newUUIDV7()),
			len(newUUIDV7().String()),
			true,
			newUUIDV7().String(),
			newUUIDV7().String(),
			newUUIDV7().String(),
			newUUIDV7().String(),
			"crypt/rand",
			"v7: 16 bytes : 8 bytes time+sequence, random with version & variant embedded",
		},
		{
			"[oklog/ulid](https://github.com/oklog/ulid)",
			len(newUlid()),
			len(newUlid().String()),
			true,
			newUlid().String(),
			newUlid().String(),
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
			betterguid.New(),
			betterguid.New(),
			"counter",
			"8 byte ts(ms) : 9 byte counter random init per ts(ms)",
		},
	}

	fmt.Printf("| Package                                                   |BLen|ELen| K-Sort| Encoded ID and Next | Method | Components |\n")
	fmt.Printf("|-----------------------------------------------------------|----|----|-------|---------------------|--------|------------|\n")

	for _, v := range packages {
		fmt.Printf("| %-57s | %d | %d | %5v | `%s`<br>`%s`<br>`%s`<br>`%s`  | %s | %s |\n",
			v.name, v.blen, v.elen, v.ksortable, v.sample, v.next, v.next2, v.next3, v.uniq, v.components)
	}
}

// ulid is configured here to be similar (random component) to rid, ksuid, uuid
func newUlid() ulid.ULID {
	t := time.Now().UTC()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	return ulid.MustNew(ulid.Timestamp(t), entropy)
}

func newUUIDV7() uuid.UUID {
	r, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	return r
}
