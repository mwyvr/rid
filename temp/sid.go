package main

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/speps/go-hashids"
)

/*
String() byte6:
zzzzzzzzzw------^@^@^@^@
String() byte7:
zzzzzzzzzzzg----^@^@^@^@
*/

func main() {
	// for i := 0; i < 10; i++ {
	// 	fmt.Printf("Processed: %d per ms\n", overload2())
	// }
	start := time.Date(2020, time.January, 0, 0, 0, 0, 0, time.UTC)
	// fmt.Printf("Unix + UnixNano from Time:\n%d\n%d\n", start.Unix(), start.UnixNano())
	id := NewWithTime(start)
	fmt.Printf("Jan 1: %3v %s len: %d %v %v counter: %d\n", id[:], id.String(), len(id.String()), id.Timestamp(), id.Time(), id.Count())
	enc := id.String()
	id, err := FromString(enc)
	if err != nil {
		panic(err)
	}
	fmt.Printf("    >: %3v %s len: %d %v %v counter: %d\n", id[:], id.String(), len(id.String()), id.Timestamp(), id.Time(), id.Count())
	fmt.Println("------------------")
	// fmt.Printf("Encoded: %s Decoded: %d %d", id, id.Timestamp(), id.Count())
	// fmt.Println("------------------")
	id = New()
	fmt.Printf("Now  : %3v %s len: %d %v %v counter: %d\n", id[:], id.String(), len(id.String()), id.Timestamp(), id.Time(), id.Count())

	fmt.Println("------------------")
	timelook()
}

type ID [rawLen]byte

const (
	rawLen     = 8
	encodedLen = 13 // testing

	// Crockford Base32 character set: i, o, l, u removed; xxyz added;
	// further modified by moving digits last.
	charset = "abcdefghjkmnpqrstvwxyz0123456789" // mod-crockford
	// charset = "0123456789abcdefghjkmnpqrstvwxyz" // crockford
	// charset = "123456789abcdfghjklmnpqrstuvwxyz" // 32 characters long only 1 vowel: a
	// charset = "0123456789abcdefghijklmnopqrstuv" // standard Base32
	// charset = "abcdefghijklmnopqrstuv0123456789" // mod-std Base32
	// charset = "ybndrfg8ejkmcpqxot1uwisza345h769" // Z-Base-32
)

var (
	// Encoding is a customized Base32 variant utlizing the Crockford character set.
	Encoding = base32.NewEncoding(charset).WithPadding(-1)
	// Encoding = base32.StdEncoding.WithPadding(-1)
	// AltEncoding is an experimental HashID encoder/decoder
	AltEncoding = codecMust()

	// counter is atomically updated, max 65535 per milliscond or 65,535,000 per second before collision.
	// counter = uint32(0) // testing
	counter = randInt()

	// ErrInvalidID is returned when trying to unmarshal an invalid ID
	ErrInvalidID = errors.New("sid: invalid ID")
)

func New() ID {
	return NewWithTime(time.Now())
}

func newWithTS(ms uint64) ID {
	var id ID
	// Big Endian
	// 6 bytes of time, to millisecond
	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)
	// 2 byte counter - unsigned int with maximum 65535
	ct := atomic.AddUint32(&counter, 1)
	id[6] = byte(ct >> 8)
	id[7] = byte(ct)

	return id
}
func NewWithTime(tm time.Time) ID {
	var id ID
	ms := Timestamp(tm)
	// Big Endian
	// 6 bytes of time, to millisecond
	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)
	// 2 byte counter - unsigned int with maximum 65535
	ct := atomic.AddUint32(&counter, 1)
	id[6] = byte(ct >> 8)
	id[7] = byte(ct)

	return id
}

func (id ID) String() string {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return *(*string)(unsafe.Pointer(&text))
}

// Time returns the timestamp component as a Go time value.
func (id ID) Time() time.Time {
	ms := id.Timestamp()
	// 1e3 / 1e6 are float constants
	s := int64(ms / 1e3)
	ns := int64((ms % 1e3) * 1e6)
	return time.Unix(s, ns)
}

// Timestamp returns the ID timestamp as the number of milliseconds from the Unix epoch.
// TODO add notes on conversion for use with the time package.
func (id ID) Timestamp() uint64 {

	return uint64(id[5]) |
		uint64(id[4])<<8 |
		uint64(id[3])<<16 |
		uint64(id[2])<<24 |
		uint64(id[1])<<32 |
		uint64(id[0])<<40

}

// Count returns the count component of the ID.
func (id ID) Count() uint16 {
	// Big-endian 2-byte value 0-65535
	return uint16(id[6])<<8 | uint16(id[7])
}

// AsHash returns ID as a random-looking variable length (10-16 byte)
// HashID representation. EXPERIMENTAL and likely going away.
func (id ID) AsHash() string {
	str, err := AltEncoding.EncodeInt64([]int64{int64(id.Timestamp()), int64(id.Count())})
	if err != nil {
		panic(err)
	}
	return str
}

// ------------------------------------------------------------------------
// FromString decodes the Base32 string and returns an ID, if possible,
func FromString(str string) (ID, error) {
	id := &ID{}
	err := id.UnmarshalText([]byte(str))
	return *id, err
}

// FromBytes convert the byte array representation of `ID` back to `ID`
func FromBytes(b []byte) (ID, error) {
	var id ID
	// TODO put back after development
	// if len(b) != rawLen {
	// 	return nilID, ErrInvalidID
	// }
	copy(id[:], b)
	return id, nil
}

// UnmarshalText implements encoding/text TextUnmarshaler interface
func (id *ID) UnmarshalText(text []byte) error {
	fmt.Printf("UnmarshalText: '%s'\n", text)
	if len(text) != encodedLen {
		fmt.Printf("huh? %#v %d\n", text, len(text))
		return ErrInvalidID
	}
	buf := make([]byte, rawLen)
	count, err := decode(buf, text)
	if (count != rawLen) || (err != nil) {
		fmt.Printf("huh? %v %d count:%d\n", buf, count)
		return ErrInvalidID
	}
	copy(id[:], buf)
	return nil
}

// encode returns the Base32 representation the supplied []byte value (id[:])
func encode(dst, id []byte) {
	Encoding.Encode(dst, id[:])
}

// decode Base32 text, returning bytes decoded and error
func decode(buf []byte, src []byte) (int, error) {
	return Encoding.Decode(buf, src)
}

// Now is a convenience function that returns the current
// UTC time in Unix milliseconds. Equivalent to:
//   Timestamp(time.Now().UTC())
func Now() uint64 { return Timestamp(time.Now().UTC()) }

// Timestamp converts a time.Time to Unix milliseconds.
//
// Because of the way ULID stores time, times from the year
// 10889 produces undefined results.
func Timestamp(t time.Time) uint64 {

	return uint64(t.Unix())*1000 +
		uint64(t.Nanosecond()/int(time.Millisecond))

}

// Time converts Unix milliseconds in the format
// returned by the Timestamp function to a time.Time.
func Time(ms uint64) time.Time {

	// 1e3 / 1e6 are float constants
	s := int64(ms / 1e3)
	ns := int64((ms % 1e3) * 1e6)
	return time.Unix(s, ns)

}

// randInt generates a random uint32
func randInt() uint32 {
	b := make([]byte, 3)
	if _, err := rand.Reader.Read(b); err != nil {
		panic(fmt.Errorf("sid: cannot generate random number: %v;", err))
	}
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
}

// codecMust configures and creates the HashID generator.
// EXPERIMENTAL, likely going away.
// In the unlikely event this function fails, a panic will result.
func codecMust() *hashids.HashID {
	c, err := hashids.NewWithData(&hashids.HashIDData{
		Alphabet:  charset,
		Salt:      "climb the mountains and get their good tidings", // we aren't trying to secure anything here
		MinLength: 10})
	if err != nil {
		panic(fmt.Errorf("sid: %s", err))
	}
	return c
}

/* ------------------------------------------------------------------------------------------------------------- */
func overload2() int {
	count := 0
	// get going on a new MS
	last := Timestamp(time.Now())
	for {
		ts := Timestamp(time.Now())
		if ts > last {
			break
		}
		_ = newWithTS(ts)
		count++
	}
	return count
}
func overload() int {
	start := Timestamp(time.Now())
	count := 0
	for {
		// get going on a new MS
		now := Timestamp(time.Now())
		if now == start {
			continue
		}
		// begin!
		last := now
		for now == last {
			t := time.Now()
			ts := Timestamp(t)
			if ts > last {
				break
			}
			_ = NewWithTime(t)
			count++
		}
		return count
	}
}

func timelook() {
	now := time.Now()
	for i := 0; i < 100; i++ {
		// time.Sleep(1 * time.Second)
		id := NewWithTime(now.AddDate(i+5, i+1, i+2))
		fmt.Printf("%3v %s len:%d %v len:%d %v %v counter: %d\n", id[:], id.String(), len(id.String()), id.AsHash(), len(id.AsHash()), id.Timestamp(), id.Time(), id.Count())
	}
}
