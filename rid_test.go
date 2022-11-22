package rid

// TODO add chronological sorting test

import (
	"bytes"
	"database/sql/driver"
	enc "encoding"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

var (
	// testing concurrency safety
	wg            sync.WaitGroup
	numConcurrent = 2000 // go routines X
	numIter       = 500  // id creation/routine
)

type idTest struct {
	name         string
	valid        bool
	id           ID
	rawBytes     []byte
	seconds      int64
	entropy      uint32
	b32          string
}

// TODO add date values in for direct comparison
var testIDS = []idTest{
	{
		"nilID",
		false,
		nilID,
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		0,
		0,
		"0000000000000000",
	},
	{
		// epoch time plus a counter of one to avoid being
		// equal to nilID, which is far as counter should never
		// be 0
		"min value 1970-01-01 00:00:00 +0000 UTC",
		true,
		ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		0,
		1,
		"0000000000002",
	},
	{
		"max value in the year 10889 see you then",
		true,
		ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		281474976710655,
		65535,
		"zzzzzzzzzzzzy",
	},
	{
		"fail on FromString / FromBytes / decode - value mismatch",
		false,
		ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xaa},
		281474976710655,
		65535,
		"1234567890abc",
	},
	{
		"fail on FromString, FromBytes len mismatch",
		false,
		ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xaa},
		281474976710655,
		65535,
		"zzzz",
	},
	{
		"must fail MarshalText (decode test - invalid base32 chars)",
		false,
		ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF},
		0,
		1,
		"zzzuzzzzzzzzt",
	},
}


func TestNew(t *testing.T) {
	id := New()
	if id == nilID {
		t.Errorf("New() produced a nilID")
	}
}

func TestNewWithTime(t *testing.T) {
	// must match
	id := NewWithTime(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
	if id.String() != "05qnwsq800002" {
		t.Errorf("ID.NewWithTime().String() not matching got %v, want %v",
			id.String(), "05qnwsq800002")
	}
}

func TestID_IsNil(t *testing.T) {
	id := New()
	if id.IsNil() {
		t.Errorf("ID.IsNil() returned %v, want %v", id.IsNil(), false)
	}
	id = ID{}
	if !id.IsNil() {
		t.Errorf("ID.IsNil() returned %v, want %v", id.IsNil(), false)
	}
}

func TestID_Seconds(t *testing.T) {
	id := NewWithTime(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
	if m := id.Seconds(); m != int64(1577836800000) {
		t.Errorf("ID.Seconds() got %v want %v", m, 1577836800000)
	}
}

func TestID_Bytes(t *testing.T) {
	id, err := FromString("05yykgvzqfzzy")
	if err != nil {
		t.Error(err)
	}
	want := []byte{1, 125, 233, 195, 127, 187, 255, 255}
	if b := id.Bytes(); bytes.Equal(b, want) != true {
		t.Errorf("ID.Bytes() got %v want %v", b, want)
	}
}

func TestID_Components(t *testing.T) {
	// for completeness
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.String(); (got != tt.b32) && (tt.valid != false) {
				t.Errorf("ID.String() = %v, want %v", got, tt.b32)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Seconds(); (got != tt.seconds) && (tt.valid != false) {
				t.Errorf("ID.Seconds() = %v %v, want %v", got, tt.id[:], tt.seconds)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Entropy(); (got != tt.entropy) && (tt.valid != false) {
				t.Errorf("ID.Entropy() = %v, want %v", got, tt.entropy)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Bytes(); !reflect.DeepEqual(got, tt.id[:]) {
				t.Errorf("ID.Bytes() = %#v, want %#v", got, tt.id)
			}
		})
	}
}

func TestID_Time(t *testing.T) {
	id, err := FromString("0000000000000")
	date := id.Time().UTC()
	if err != nil {
		t.Error(err)
	}
	y, m, d := date.Date()
	if (y != 1970) || (m != 1) || (d != 1) {
		t.Errorf("ID.Time() returned %d, %d, %d; want 1970,1,1", y, m, d)
	}
	// now
	id = NewWithTime(time.Now())
	if int64(id.Time().UnixNano()/1e6) != id.Seconds() {
		t.Errorf("ID.Time() UnixNano()/1e6 != id.Seconds")
	}
}

func TestFromString(t *testing.T) {
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromString(tt.b32)
			if tt.valid && (err != nil) {
				t.Errorf("FromString() error = %v, is valid %v", err, tt.valid)
				return
			}
			_ = err
			if tt.valid && !reflect.DeepEqual(got, tt.id) {
				t.Errorf("FromString() = %v, want %v", got, tt.id)
			}
		})
	}
	// callers should lowercase.
	got, err := FromString("aaaaaaaaaaaaA")
	if err == nil {
		t.Errorf("Should be an error")
	} else if err != ErrInvalidID {
		t.Errorf("FromString() = %v, want err %v got %v", got, err, ErrInvalidID)
	}
	// decoding the nilID value is legit
	got, err = FromString("aaaaaaaaaaaaa")
	if err != nil {
		t.Errorf("FromString(\"aaaaaaaaaaaaa\") nilID value failed, got %v, %v", got, err)
	}
}

func TestFromBytes(t *testing.T) {
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromBytes(tt.rawBytes)
			if tt.valid && (err != nil) {
				t.Errorf("FromBytes() error = %v, wantErr %v", err, tt.valid)
				return
			}
			if tt.valid && !reflect.DeepEqual(got, tt.id) {
				t.Errorf("FromBytes() = %v, want %v", got, tt.id)
			}
		})
	}
	// nilID byte value is unusual but legit
	got, err := FromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		t.Errorf("FromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) nilID value failed, got %v, %v", got, err)
	}
	// invalid len
	got, err = FromBytes([]byte{0x12, 0x34})
	if err == nil {
		t.Errorf("FromBytes([]byte{0x12, 0x34}) got %v, err==nil, %v", got, err)
	}
}

func Test_encode(t *testing.T) {
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			dest := make([]byte, encodedLen)
			encode(dest, tt.id[:])
			if len(dest) != encodedLen {
				t.Errorf("encode: wrong length got %d want %d", len(dest), encodedLen)
				return
			}
			if tt.valid && string(dest) != tt.b32 {
				t.Errorf("encode: wrong output got %s want %s", dest, tt.b32)
			}
		})
	}
}

func Test_decode(t *testing.T) {
	id := &ID{}
	// there really are no checks in decode; they happen in UnmarshalText,
	// the only caller of decode(). For code coverage:
	decode(id, []byte("05yykgvzqc002"[:]))
	if id.Entropy() != 1 {
		t.Errorf("decode produced an anomoly: %#v", id)
	}
}

func TestID_UnmarshalText(t *testing.T) {
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure that ID fulfills the Interface
			var _ enc.TextUnmarshaler = &ID{}
			text := []byte(tt.b32[:])
			if err := tt.id.UnmarshalText(text); err != nil {
				if tt.valid { // shouldn't be
					t.Errorf("ID.UnmarshalText() error = %v, want %v", err, tt.id[:])
				}
				if !tt.valid && err != ErrInvalidID {
					t.Errorf("ID.UnmarshalText() error = %v, want %v", err, ErrInvalidID)
				}
			}
		})
	}
}

func TestID_MarshalText(t *testing.T) {
	// ensure ID implements the interface
	var _ enc.TextMarshaler = &ID{}
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.id.MarshalText()
			if tt.valid && (err != nil) {
				t.Errorf("ID.MarshalText() error = %v, wantErr %v", err, tt.valid)
				return
			}
			if tt.valid && string(got) != tt.b32 {
				t.Errorf("ID.MarshalText() = %v, want %v", got, tt.b32)
			}
		})
	}
}

func TestID_Value(t *testing.T) {
	// ensure ID implements the interface
	var _ driver.Valuer = &ID{}
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.id.Value()
			if tt.valid && (err != nil) {
				t.Errorf("ID.Value() error = %v, is valid %v", err, tt.valid)
				return
			}
			if tt.valid && !reflect.DeepEqual(got, tt.b32) {
				t.Errorf("ID.Value() = %v, want %v test: %v", got, tt.b32, tt)
			}
		})
	}
	// nilID
	val, err := nilID.Value()
	if (val != nil) && (err != nil) {
		t.Errorf("ID.Value(nilID) want nil, nil, got %v, %v", val, err)
	}
}

func TestID_Scan(t *testing.T) {
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			// Scan string
			if err := tt.id.Scan(tt.b32); tt.valid && (err != nil) {
				t.Errorf("ID.Scan() error = %v, valid? %v", err, tt.valid)
			}
			// Scan []byte
			bs := []byte(tt.b32)
			if err := tt.id.Scan(bs); tt.valid && (err != nil) {
				t.Errorf("ID.Scan() error = %v, valid? %v", err, tt.valid)
			}
		})
	}
	// nil
	id := New()
	if err := id.Scan(nil); err != nil {
		t.Errorf("ID.Scan() error = %v, should return nilID", err)
	}
	if bytes.Equal(id[:], nilID[:]) != true {
		t.Errorf("ID.Scan() got %v, should return nilID %v", id, nilID)
	}
	// unsupported
	id = ID{}
	if err := id.Scan(false); err == nil {
		t.Errorf("ID.Scan() error = %v, should not convert bool", err)
	}
}


// Benchmarks
func BenchmarkIDNew(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next(){
            _ = New()
        }
    })
}

func BenchmarkIDNewEncoded(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next(){
            _ = New().String()
        }
    })
}

// examples
func ExampleNew() {
	id := New()
	fmt.Printf(`ID:
    String()       %s   
    Seconds() %d  
    Entropy()        %d // random for this one-off run 
    Time()         %v
    Bytes():       %3v  
`, id.String(), id.Seconds(), id.Entropy(), id.Time(), id.Bytes())
}

func ExampleNewWithTime() {
	id := NewWithTime(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
	fmt.Printf(`ID:
    String()       %s
    Seconds() %d
    Entropy()        %d // random for this one-off run 
    Time()         %v
    Bytes():       %3v
`, id.String(), id.Seconds(), id.Entropy(), id.Time().UTC(), id.Bytes())
}

func ExampleFromString() {
	id, err := FromString("05yx13hj9kq4g")
	if err != nil {
		panic(err)
	}
	fmt.Println(id.Seconds(), id.Entropy())
	// [05yx13hj9kq4g] ms:1639881519692 count:61000 time:2021-12-18 18:38:39.692 -0800 PST id:{1, 125, 208, 142, 50, 76, 238, 72}
}

func TestID_MarshalJSON(t *testing.T) {
	if got, err := nilID.MarshalJSON(); string(got) != "null" {
		t.Errorf("ID.MarshalJSON() of nilID error = %v, got %v", err, got)
	}
	if got, err := (ID{1, 125, 208, 142, 50, 76, 238, 72}).MarshalJSON(); string(got) != "\"05yx13hj9kq4g\"" {
		if err != nil {
			t.Errorf("ID.MarshalJSON() err %v marshaling %v", err, "\"05yx13hj9kq4g\"")
		}
		t.Errorf("ID.MarshalJSON() got %v want %v", string(got), "\"05yx13hj9kq4g\"")
	}
}

func TestID_UnmarshalJSON(t *testing.T) {
	var id ID
	err := id.UnmarshalJSON([]byte("null"))
	if err != nil {
		t.Errorf("ID.UnmarshalJSON() of null, error = %v", err)
	}
	if id != nilID {
		t.Errorf("ID.UnmarshalJSON() error = %v", err)
	}
	// 2020...
	text := []byte("\"05yykgvzqc002\"")
	if err = id.UnmarshalJSON(text); err != nil {
		t.Errorf("ID.UnmarshalJSON() error = %v", err)

	} else if id != (ID{1, 125, 233, 195, 127, 187, 0, 1}) {
		t.Errorf("ID.UnmarshalJSON() of %v, got %v", text, id.String())
	}
}
