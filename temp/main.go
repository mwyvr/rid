package main

import (
	"fmt"

	"github.com/solutionroute/rid"
)

type idParts struct {
	id        rid.ID
	encoded   string
	encoded64 string
	ts        int64
	rtsig     []byte
	random    uint64
}

var IDs = []idParts{
	// sorted (ascending) should be IDs 2, 1, 3, 0
	// 	062ektdeb6039z5masctt333 AYTp6a5ZgDT8tFZZrQxj
	// zzzzzzzzzzzzzzzzzzzg0000 ________________AAAA
	// 000000000000000000000000 AAAAAAAAAAAAAAAAAAAA
	// 062ektcmm0k3bgwxd4bceqtb AYTp6ZSgJjXDnWkWx19L

	{
		// 062ektdeb6039z5masctt333 ts:1670371716697 rtsig:[0x80] random: 58259961960877 | time:2022-12-06 16:08:36.697 -0800 PST ID{0x1,0x84,0xe9,0xe9,0xae,0x59,0x80,0x34,0xfc,0xb4,0x56,0x59,0xad,0xc,0x63}
		rid.ID{0x1, 0x84, 0xe9, 0xe9, 0xae, 0x59, 0x80, 0x34, 0xfc, 0xb4, 0x56, 0x59, 0xad, 0xc, 0x63},
		"062ektdeb6039z5masctt333",
		"AYTp6a5ZgDT8tFZZrQxj",
		1670371716697,
		[]byte{0x80},
		58259961960877,
	},
	{
		rid.ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		"zzzzzzzzzzzzzzzzzzzg0000",
		"________________AAAA",
		0,
		[]byte{0x00},
		0,
	},
	{
		rid.ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		"000000000000000000000000",
		"AAAAAAAAAAAAAAAAAAAA",
		0,
		[]byte{0x00},
		0,
	},
	{
		// 062ektcmm0k3bgwxd4bceqtb ts:1670371710112 rtsig:[0x26] random: 59114275804871 | time:2022-12-06 16:08:30.112 -0800 PST ID{0x1,0x84,0xe9,0xe9,0x94,0xa0,0x26,0x35,0xc3,0x9d,0x69,0x16,0xc7,0x5f,0x4b}
		rid.ID{0x1, 0x84, 0xe9, 0xe9, 0x94, 0xa0, 0x26, 0x35, 0xc3, 0x9d, 0x69, 0x16, 0xc7, 0x5f, 0x4b},
		"062ektcmm0k3bgwxd4bceqtb",
		"AYTp6ZSgJjXDnWkWx19L",
		1670371710112,
		[]byte{0x26},
		59114275804871,
	},
}

func main() {

	for _, v := range IDs {
		id, err := rid.FromBytes(v.id[:])
		if err != nil {
			panic(err)
		}
		// fmt.Println(id.String(), rid.String64(id))
		fmt.Printf("%s %-15d %40s %d\n", id, id.Timestamp(), id.Time(), id.Random())
	}

}
