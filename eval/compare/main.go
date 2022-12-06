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
			true,
			rid.New().String(),
			rid.New().String(),
			"fastrand",
			"ts(seconds) : runtime signature : random",
		},
		{
			"[rs/xid](https://github.com/rs/xid)",
			len(xid.New().Bytes()),
			len(xid.New().String()),
			true,
			true,
			xid.New().String(),
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
			betterguid.New(),
			"counter",
			"ts(ms) + per-ms math/rand initialized counter",
		},
	}

	fmt.Printf("| Package                                                   |BLen|ELen| K-Sort| 0-Cfg | Encoded ID and Next | Method | Components |\n")
	fmt.Printf("|-----------------------------------------------------------|----|----|-------|-------|---------------------|--------|------------|\n")
	for _, v := range packages {
		fmt.Printf("| %-57s | %d | %d | %5v | %5v | `%s`<br>`%s` | %s | %s |\n",
			v.name, v.blen, v.elen, v.ksortable, v.zeroconfig, v.sample, v.next, v.uniq, v.components)
	}
	// t := time.Now()
	// fmt.Println(t.Unix())
	// fmt.Println(t.UnixMilli())
	// fmt.Println(t.UnixNano())
	// // nano := uint32(1670361308664)
	// bs := []byte(strconv.Itoa(int(t.Unix())))
	// fmt.Println(bs)
	//
	// buf := new(bytes.Buffer)
	// err := binary.Write(buf, binary.BigEndian, t.UnixMilli())
	// if err != nil {
	// 	fmt.Println("binary.Write failed:", err)
	// }
	// fmt.Printf("Time: % x\n", buf.Bytes())
	// fmt.Printf("rid: % x\n", rid.New().Bytes())

}

// ulid is configured here to be similar (random component) to rid, ksuid, uuid
func newUlid() ulid.ULID {
	t := time.Now().UTC()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	return ulid.MustNew(ulid.Timestamp(t), entropy)
}
