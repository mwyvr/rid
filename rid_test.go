// Acknowledgement: This source file is based on work in package github.com/rs/xid,
package rid

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

type idParts struct {
	id      ID
	encoded string
	ts      int64
	seq     int64
	random  uint64
}

var CHECKIDS = []idParts{
	// sorted (ascending) should be IDs 2, 3, 0, 5, 4, 1
	{
		// 03f6nlxczw000000 ts:946684799999 seq:   0 rnd:    0 1999-12-31 23:59:59.999 +0000 UTC ID{  0x0, 0xdc, 0x6a, 0xcf, 0xab, 0xff,  0x0,  0x0,  0x0,  0x0 }
		ID{0x0, 0xdc, 0x6a, 0xcf, 0xab, 0xff, 0x0, 0x0, 0x0, 0x0},
		"03f6nlxczw000000",
		946684799999,
		0,
		0,
	},
	{
		// zzzzzzzzzzzzzzzz ts:281474976710655 seq:65535 rnd:65535 10889-08-02 05:31:50.655 +0000 UTC ID{ 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff }
		ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		"zzzzzzzzzzzzzzzz",
		281474976710655,
		65535,
		65535,
	},
	{
		// 0000000000000000 ts:0 seq:   0 rnd:    0 1970-01-01 00:00:00 +0000 UTC ID{  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0,  0x0 }
		ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		"0000000000000000",
		0,
		0,
		0,
	},
	{
		// 02k4he6ej8000t4f ts:696996122002 seq:   0 rnd:26766 1992-02-02 02:02:02.002 +0000 UTC ID{  0x0, 0xa2, 0x48, 0x34, 0xcd, 0x92,  0x0,  0x0, 0x68, 0x8e }
		ID{0x0, 0xa2, 0x48, 0x34, 0xcd, 0x92, 0x0, 0x0, 0x68, 0x8e},
		"02k4he6ej8000t4f",
		696996122002,
		0,
		26766,
	},
	{
		// 06bpjb8pz0000000 ts:1741226055416 seq:   0 rnd:    0 2025-03-05 17:54:15.416 -0800 PST ID{  0x1, 0x95, 0x69, 0x29, 0x16, 0xf8,  0x0,  0x0,  0x0,  0x0 }
		ID{0x1, 0x95, 0x69, 0x29, 0x16, 0xf8, 0x0, 0x0, 0x0, 0x0},
		"06bpjb8pz0000000",
		1741226055416,
		0,
		0,
	},
	{
		// 05z169vrs40006zf ts:1640998861001 seq:   0 rnd: 7150 2022-01-01 01:01:01.001 +0000 UTC ID{  0x1, 0x7e, 0x13, 0x27, 0x78, 0xc9,  0x0,  0x0, 0x1b, 0xee }
		ID{0x1, 0x7e, 0x13, 0x27, 0x78, 0xc9, 0x0, 0x0, 0x1b, 0xee},
		"05z169vrs40006zf",
		1640998861001,
		0,
		7150,
	},
}

func TestIDPartsExtraction(t *testing.T) {
	for i, v := range CHECKIDS {
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
			if got, want := v.id.Time(), time.UnixMilli(v.ts).UTC(); got != want {
				t.Errorf("Time() = %v, want %v", got, want)
			}
			if got, want := v.id.Timestamp(), v.ts; got != want {
				t.Errorf("Time() = %v, want %v", got, want)
			}
			if got, want := v.id.Random(), v.random; got != want {
				t.Errorf("Random() = %v, want %v", got, want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	// Generate N ids, see if all unique
	// Parallel generation test is in ./cmd/eval/uniqcheck/main.go
	count := 10000
	ids := make([]ID, count)
	for i := range count {
		ids[i] = New()
	}
	for i := 1; i < count; i++ {
		prevID := ids[i-1]
		id := ids[i]
		// Test for uniqueness among all other generated ids
		for j, tid := range ids {
			if j != i {
				// can't use ID.Compare for this test as it compares only the time
				// component of IDs
				if bytes.Equal(id[:], tid[:]) {
					t.Errorf("generated ID is not unique (%d/%d)\n%v", i, j, ids)
				}
			}
		}
		// Check that timestamp was incremented and is within 30 seconds (30000 ms) of the previous one
		secs := id.Time().Sub(prevID.Time()).Seconds()
		if secs < 0 || secs > 30 {
			t.Error("wrong timestamp in generated ID")
		}
	}
}

// ensure sequencing produces unique ts+seq combos
func TestSequence(t *testing.T) {
	var lastTS, lastSeq int64
	// Generate 1,000,000 new IDs on the fly (about 70 ms depending on hardware)
	check := []ID{}
	for range 1000000 {
		check = append(check, New())
	}
	for _, id := range check {
		if lastTS != id.Timestamp() {
			lastTS = id.Timestamp()
			lastSeq = id.Sequence()
			continue
		}
		if id.Timestamp() == lastTS && id.Sequence() <= lastSeq {
			t.Errorf("sequence not unique for next ID ts: %d seq: %d last: %d", id.Timestamp(), id.Sequence(), lastTS)
		} else {
			lastSeq = id.Sequence()
		}
	}
}

func TestIDString(t *testing.T) {
	for _, v := range CHECKIDS {
		if got, want := v.encoded, v.id.String(); got != want {
			t.Errorf("String() = %v, want %v", got, want)
		}
	}
}

func TestIDEncode(t *testing.T) {
	id := ID{0x63, 0xac, 0x76, 0xd3, 0xff, 0xff, 0xfc, 0x30, 0x37, 0xc2}
	text := make([]byte, encodedLen)
	if got, want := string(id.Encode(text)), "dfp7emzzzzy30ey2"; got != want {
		t.Errorf("Encode() = %v, want %v", got, want)
	}
}

func TestFromString(t *testing.T) {
	// 06bprdfln4x281hd ts:1741276959657 seq:14884 rnd: 1548 2025-03-06 16:02:39.657 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x31, 0xd3, 0xa9, 0x3a, 0x24,  0x6,  0xc }
	got, err := FromString("06bprdfln4x281hd")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x95, 0x6c, 0x31, 0xd3, 0xa9, 0x3a, 0x24, 0x6, 0xc}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
	// nil ID
	got, err = FromString("0000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
	// max ID
	got, err = FromString("zzzzzzzzzzzzzzzz")
	if err != nil {
		t.Fatal(err)
	}
	want = ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	if got != want {
		t.Errorf("FromString() = %v, want %v", got, want)
	}
}

func TestFromStringInvalid(t *testing.T) {
	_, err := FromString("012345")
	if err != ErrInvalidID {
		t.Errorf("FromString(invalid length) err=%v, want %v", err, ErrInvalidID)
	}
	id, err := FromString("062ez870acdtzd2y3qajilou") // i, l, o, u never in our IDs
	if err != ErrInvalidID {
		t.Errorf("FromString(062ez870acdtzd2y3qajilou - invalid chars) err=%v, want %v", err, ErrInvalidID)
	}
	if id != nilID {
		t.Errorf("FromString() =%v, there want %v", id, nilID)
	}
}

func TestID_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		id      *ID
		wantErr bool
	}{
		{
			// 0000000000000000 ts:0 rnd:              0 1969-12-31 16:00:00 -0800 PST ID{0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0,0x0}
			"valid",
			"0000000000000000",
			&nilID,
			false,
		},
		{
			"invalid chars",
			"000000000000000u",
			&nilID,
			true,
		},
		{
			"invalid length too long",
			"12345678901",
			&nilID,
			true,
		},
		{
			"invalid length too short",
			"dfb7emm",
			&nilID,
			true,
		},
		{
			// 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
			"valid id",
			"06bprg666xzm7hpg",
			&ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.id.UnmarshalText([]byte(tt.text)); (err != nil) != tt.wantErr {
				t.Errorf("ID.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestID_UnmarshalTextError(t *testing.T) {
	id := nilID
	if err := id.UnmarshalText([]byte("invalid")); err != ErrInvalidID {
		t.Errorf("ID.UnmarshalText() error = %v, wantErr %v", err, ErrInvalidID)
	}
	id = New() // make a non nil ID
	if err := id.UnmarshalText([]byte("foo")); id != nilID {
		t.Errorf("ID.UnmarshalText() want nil ID, ErrInvalidID, got %v, %v", id, err)
	}
}

func TestID_IsNil(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want bool
	}{
		{
			name: "ID not nil",
			id:   New(),
			want: false,
		},
		{
			name: "Nil ID",
			id:   ID{},
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got, want := tt.id.IsNil(), tt.want; got != want {
				t.Errorf("IsNil() = %v, want %v", got, want)
			}
		})
	}
}

// is an alias function, no need to repeat above, just for coverage report
func TestID_IsZero(t *testing.T) {
	id := ID{}
	if !id.IsZero() {
		t.Errorf("ID.IsZero() = %v, want %v", id.IsZero(), true)
	}
}

func TestNilID(t *testing.T) {
	got := ID{}
	if want := NilID(); !reflect.DeepEqual(got, want) {
		t.Error("NilID() not equal ID{}")
	}
}

func TestNilID_IsNil(t *testing.T) {
	if !NilID().IsNil() {
		t.Error("NilID().IsNil() is not true")
	}
}

func TestFromBytes_Invariant(t *testing.T) {
	want := New()
	got, err := FromBytes(want.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got[:], want[:]) {
		t.Error("FromBytes(id.Bytes()) != id")
	}
	// invalid
	got, err = FromBytes([]byte{0x1, 0x2})
	if !bytes.Equal(got[:], nilID[:]) {
		t.Error("FromBytes([]byte{0x1, 0x2}) - invalid - != nilID")
	}
	if err == nil {
		t.Fatal(err)
	}
}

type jsonType struct {
	ID  *ID
	Str string
}

func TestIDJSONMarshaling(t *testing.T) {
	// 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
	id := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	v := jsonType{ID: &id, Str: "valid"}
	data, err := json.Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(data), `{"ID":"06bprg666xzm7hpg","Str":"valid"}`; got != want {
		t.Errorf("json.Marshal() = %v, want %v", got, want)
	}
}

func TestIDJSONUnmarshaling(t *testing.T) {
	// 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
	data := []byte(`{"ID":"06bprg666xzm7hpg","Str":"valid"}`)
	v := jsonType{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	if got := *v.ID; !bytes.Equal(got[:], want[:]) {
		t.Errorf("json.Unmarshal() = %v, want %v", got, want)
	}
	// should not fail
	err = json.Unmarshal([]byte(`null`), &v)
	if err != nil {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, nil)
	}
}

func TestIDJSONUnmarshalingError(t *testing.T) {
	v := jsonType{}
	// callers are responsible for forcing lower case input for Base32
	// otherwise valid id:
	err := json.Unmarshal([]byte(`{"ID":"06BPRG666XZM7HPG"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	// too short
	err = json.Unmarshal([]byte(`{"ID":"06bprg666xzm"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	// no 'a' in character set
	err = json.Unmarshal([]byte(`{"ID":"0000000000000a"}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
	// invalid on multiple levels
	err = json.Unmarshal([]byte(`{"ID":1}`), &v)
	if err != ErrInvalidID {
		t.Errorf("json.Unmarshal() err=%v, want %v", err, ErrInvalidID)
	}
}

func TestIDDriverValue(t *testing.T) {
	// 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
	id := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	got, err := id.Value()
	if err != nil {
		t.Fatal(err)
	}
	if want := "06bprg666xzm7hpg"; got != want {
		t.Errorf("Value() = %v, want %v", got, want)
	}
}

func TestIDDriverScan(t *testing.T) {
	// 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
	got := ID{}
	err := got.Scan("06bprg666xzm7hpg")
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	if !bytes.Equal(got[:], want[:]) {
		t.Errorf("Scan() = %v, want %v", got, want)
	}
}

func TestIDDriverScanError(t *testing.T) {
	id := ID{}

	if got, want := id.Scan(0), errors.New("rid: scanning unsupported type: int"); got.Error() != want.Error() {
		t.Errorf("Scan() err=%v, want %v", got, want)
	}
	if got, want := id.Scan("0"), ErrInvalidID; got != want {
		t.Errorf("Scan() err=%v, want %v", got, want)
		if id != nilID {
			t.Errorf("Scan() id=%v, want %v", got, nilID)
		}
	}
}

func TestIDDriverScanByteFromDatabase(t *testing.T) {
	// 06bprg666xzm7hpg ts:1741277677111 seq:32579 rnd:49871 2025-03-06 16:14:37.111 +0000 UTC ID{  0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf }
	got := ID{}
	bs := []byte("06bprg666xzm7hpg")
	err := got.Scan(bs)
	if err != nil {
		t.Fatal(err)
	}
	want := ID{0x1, 0x95, 0x6c, 0x3c, 0xc6, 0x37, 0x7f, 0x43, 0xc2, 0xcf}
	if !bytes.Equal(got[:], want[:]) {
		t.Errorf("Scan() = %v, want %v", got, want)
	}
}

func TestFromBytes_InvalidBytes(t *testing.T) {
	cases := []struct {
		length     int
		shouldFail bool
	}{
		{rawLen - 1, true},
		{rawLen, false},
		{rawLen + 1, true},
	}
	for _, c := range cases {
		b := make([]byte, c.length)
		_, err := FromBytes(b)
		if got, want := err != nil, c.shouldFail; got != want {
			t.Errorf("FromBytes() error got %v, want %v", got, want)
		}
	}
}

func TestCompare(t *testing.T) {
	pairs := []struct {
		left     ID
		right    ID
		expected int
	}{
		{CHECKIDS[1].id, CHECKIDS[0].id, 1},
		{ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, CHECKIDS[2].id, 0},
		{CHECKIDS[0].id, CHECKIDS[0].id, 0},
		{CHECKIDS[2].id, CHECKIDS[1].id, -1},
		{CHECKIDS[5].id, CHECKIDS[4].id, -1},
	}
	for _, p := range pairs {
		if p.expected != p.left.Compare(p.right) {
			t.Errorf("%s Compare to %s should return %d", p.left, p.right, p.expected)
		}
		if -1*p.expected != p.right.Compare(p.left) {
			t.Errorf("%s Compare to %s should return %d", p.right, p.left, -1*p.expected)
		}
	}
}

var IDList = []ID{CHECKIDS[0].id, CHECKIDS[1].id, CHECKIDS[2].id, CHECKIDS[3].id, CHECKIDS[4].id, CHECKIDS[5].id}

func TestSorter_Len(t *testing.T) {
	if got, want := sorter([]ID{}).Len(), 0; got != want {
		t.Errorf("Len() %v, want %v", got, want)
	}
	if got, want := sorter(IDList).Len(), 6; got != want {
		t.Errorf("Len() %v, want %v", got, want)
	}
}

func TestSorter_Less(t *testing.T) {
	// sorted (ascending) should be IDs 2, 3, 0, 1
	sorter := sorter(IDList)
	if !sorter.Less(0, 1) {
		t.Errorf("Less(0, 1) not true")
	}
	if sorter.Less(3, 2) {
		t.Errorf("Less(2, 1) true")
	}
	if sorter.Less(0, 0) {
		t.Errorf("Less(0, 0) true")
	}
}

func TestSorter_Swap(t *testing.T) {
	ids := make([]ID, 0)
	ids = append(ids, IDList...)
	sorter := sorter(ids)
	sorter.Swap(0, 1)
	if got, want := ids[0], IDList[1]; !reflect.DeepEqual(got, want) {
		t.Error("ids[0] != IDList[1]")
	}
	if got, want := ids[1], IDList[0]; !reflect.DeepEqual(got, want) {
		t.Error("ids[1] != IDList[0]")
	}
	sorter.Swap(2, 2)
	if got, want := ids[2], IDList[2]; !reflect.DeepEqual(got, want) {
		t.Error("ids[2], IDList[2]")
	}
}

func TestSort(t *testing.T) {
	ids := make([]ID, 0)
	ids = append(ids, IDList...)
	Sort(ids)
	// sorted (ascending) should be IDs 2, 3, 0, 5, 4, 1
	if got, want := ids, []ID{IDList[2], IDList[3], IDList[0], IDList[5], IDList[4], IDList[1]}; !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot %v\nwant %v\n", got, want)
	}
}

func TestFastrand48New(t *testing.T) {
	t.Run("check-dupes", func(t *testing.T) {
		// see eval/uniqcheck/main.go for proof of utility testing in concurrent environments
		var id ID
		// since the underlying structure of ID is an array, not a slice, ID can be a key
		keys := make(map[ID]bool)
		count := 100000
		for i := 0; i < count; i++ {
			id = New()
			if keys[id] {
				// It's not actually an error but something to consider; no application using
				// this package ought to be generating 100,000 IDs a second let alone millions.
				t.Errorf("Duplicate random number %d generated within %d iterations", id, count)
			}
			keys[id] = true
		}
	})
}

// Benchmarks
var (
	// added to avoid compiler over-optimization and silly results
	benchResultID     ID
	benchResultString string
)

// Create new ID
func BenchmarkNew(b *testing.B) {
	var r ID
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = New()
		}
		benchResultID = r
	})
}

// common use case, generate an ID, encode as a string:
func BenchmarkNewString(b *testing.B) {
	var r string
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = New().String()
		}
		benchResultString = r
	})
}

// encoding performance only
func BenchmarkString(b *testing.B) {
	id := New()
	var r string
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r = id.String()
		}
		benchResultString = r
	})
}

// decoding performance only
func BenchmarkFromString(b *testing.B) {
	var r ID
	str := "06bprlcm7q4z16vh"
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r, _ = FromString(str)
		}
		benchResultID = r
	})
}

// examples
func ExampleNew() {
	id := New()
	fmt.Printf(`ID:
    String()  %s
    Timestamp() %d
    Sequence() %d
    Random()  %d 
    Time()    %v
    Bytes()   %3v\n`, id.String(), id.Timestamp(), id.Sequence(), id.Random(), id.Time().UTC(), id.Bytes())
}

func ExampleFromString() {
	id, err := FromString("03f6nlxczw0018fz")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Timestamp(), id.Random())
	// Output: 946684799999 41439
}
