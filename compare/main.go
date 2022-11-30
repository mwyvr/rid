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
	zeroconfig bool
	sample     string
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
			true,
			rid.New().String(),
			"crypt/rand",
			"ts(seconds) : machine ID : process ID : random",
		},
		{
			"[rs/xid](https://github.com/rs/xid)",
			len(xid.New().Bytes()),
			len(xid.New().String()),
			true,
			true,
			xid.New().String(),
			"counter",
			"ts(seconds) : machine ID : process ID : counter",
		},
		{
			"[segmentio/ksuid](https://github.com/segmentio/ksuid)",
			len(ksuid.New().Bytes()),
			len(ksuid.New().String()),
			true,
			true,
			ksuid.New().String(),
			"random",
			"ts(seconds) : random",
		},
		{
			"[google/uuid](https://github.com/google/uuid)",
			len(uuid.New()),
			len(uuid.New().String()),
			false,
			true,
			uuid.New().String(),
			"crypt/rand",
			"(v4) version + variant + 122 bits random",
		},
		{
			"[oklog/ulid](https://github.com/oklog/ulid)",
			len(newUlid()),
			len(newUlid().String()),
			true,
			true,
			newUlid().String(),
			"crypt/rand",
			"ts(ms) : choice of random",
		},
		{
			"[kjk/betterguid](https://github.com/kjk/betterguid)",
			8 + 12, // only available as a string
			len(betterguid.New()),
			true,
			true,
			betterguid.New(),
			"counter",
			"ts(ms) + per-ms math/rand initialized counter",
		},
	}

	fmt.Printf("| Package                                                   |BLen|ELen| K-Sort| 0-Cfg | Encoded ID                           | Method     | Components |\n")
	fmt.Printf("|-----------------------------------------------------------|----|----|-------|-------|--------------------------------------|------------|------------|\n")
	for _, v := range packages {
		fmt.Printf("| %-57s | %d | %d | %5v | %5v | %-36s | %-10s | %s |\n",
			v.name, v.blen, v.elen, v.ksortable, v.zeroconfig, v.sample, v.uniq, v.components)
	}

}

// ulid is configured here to be similar (random component) to rid, ksuid, uuid
func newUlid() ulid.ULID {
	t := time.Now().UTC()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	return ulid.MustNew(ulid.Timestamp(t), entropy)
}
