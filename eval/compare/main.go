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
			"[solutionroute/rid](https://github.com/solutionroute/rid)<br>Base32 (default)",
			len(rid.New().Bytes()),
			len(rid.New().String()),
			true,
			true,
			rid.New().String(),
			rid.New().String(),
			"fastrand",
			"6 byte ts(ms) : 1 byte machine/pid signature : 6 byte random",
		},
		{
			"[solutionroute/rid](https://github.com/solutionroute/rid)<br>Base64 (included helper functions)",
			len(rid.New().Bytes()),
			len(rid.String64(rid.New())),
			true,
			true,
			rid.String64(rid.New()),
			rid.String64(rid.New()),
			"fastrand",
			"6 byte ts(ms) : 1 byte machine/pid signature : 6 byte random",
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
			"4 byte ts(sec) : 2 byte mach ID : 2 byte pid : 3 byte monotonic counter",
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
			"4 byte ts(sec) : 16 byte random",
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
			"v4: 16 bytes random with version & variant embedded",
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
			"6 byte ts(ms) : 10 byte counter random init per ts(ms)",
		},
		{
			"[kjk/betterguid](https://github.com/kjk/betterguid)",
			8 + 9, // only available as a string
			len(betterguid.New()),
			true,
			true,
			betterguid.New(),
			betterguid.New(),
			"counter",
			"8 byte ts(ms) : 9 byte counter random init per ts(ms)",
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
