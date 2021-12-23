/*
Package sid provides a unique-enough ID generator for applications with modest
(read: non-distributed), needs.

    id := sid.New()
    fmt.Printf("%s", id) // af1zwtepacw38

A sid ID is 8-byte value; it could optionally be stored as a 64 bit integer.

    // af1zwtepacw38
    sid.FromString("af1zwtepacw38") == id   // true
    fmt.Println(id[:])                      // [1 125 227 253 59 110 47 62]

Each ID's 8-byte binary representation: id{1, 111, 89, 64, 140, 0, 165, 159} is
comprised of a:

- 6-byte timestamp value representing milliseconds since the Unix epoch
- 2-byte concurrency-safe counter (test included); maxCounter = uint16(65535)

IDs are chronologically sortable with a minor tradeoff in millisecond-level
sortability made for improved randomness in the trailing counter value.

The String() representation us base32 encoded using a modified Crockford inspired alphabet.

Acknowledgement: Much of this package is based on the globally-unique capable
rs/xid package which itself levers ideas from mongodb. See https://github.com/rs/xid.
*/
package sid

import (
	"database/sql/driver"
	"encoding/base32"
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
	"unsafe"
)

// ID represents a locally unique and random-enough yet sortable identifier
type ID [rawLen]byte

const (
	rawLen     = 8             // bytes
	encodedLen = 13            // base32 representation
	maxCounter = uint32(65535) // max IDs per millisecond, not achievable and safe

	/*
		ID string representations are base32-encoded using a character set
		inspired by Crockford's (i, o, l, u removed and w, x, y, z added). sid's
		character set has digits moved to the end  to avoid producing IDs with
		leading zeros for many years.

		encoding/Base32 for comparison:
		          "0123456789abcdefghijklmnopqrstuv"
	*/
	charset = "abcdefghjkmnpqrstvwxyz0123456789"
)

var (
	// ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	nilID ID

	counter uint32 // this is uint32 to take advantage of atomic.CompareAndSwap...
	last    uint64 // the last millisecond since Unix epoch when an id was generated

	ErrInvalidID = errors.New("sid: invalid id")

	// dec is the decoding map for base32 encoding; currently only used for error checking
	dec      [256]byte
	encoding = base32.NewEncoding(charset).WithPadding(-1)
)

func init() {
	// initialize the base32 decoding table
	for i := 0; i < len(dec); i++ {
		dec[i] = 0xFF
	}
	for i := 0; i < len(charset); i++ {
		dec[charset[i]] = byte(i)
	}

	// We don't need crypto/rand for a random-ish solution
	rand.Seed(time.Now().UnixNano())
	counter = randInt()
}

// New returns a new ID using the current time; IDs represent millisecond time resolution.
func New() ID {
	return NewWithTime(time.Now())
}

// NewWithTime returns a new ID based upon the supplied Time value.
func NewWithTime(tm time.Time) ID {
	var id ID
	// timestamp truncated to milliseconds
	var ms = uint64(tm.Unix())*1000 + uint64(tm.Nanosecond()/int(time.Millisecond))

	// Package atomic's operations are concurrency safe.
	if ms != atomic.LoadUint64(&last) {
		atomic.StoreUint64(&last, ms)           // we're in a new ms, save it
		atomic.StoreUint32(&counter, randInt()) // randomly initialize the counter
	}
	// Assemble the 8 byte ID
	// 6 bytes of time, to millisecond
	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)

	// If counter hits max uint16 value, roll over
	atomic.CompareAndSwapUint32(&counter, maxCounter, 0)
	// increment by 1
	ct := atomic.AddUint32(&counter, 1)
	// 2 bytes for the counter
	id[6] = byte(ct >> 8)
	id[7] = byte(ct)

	return id
}

// IsNil returns true if ID == nilID
func (id ID) IsNil() bool {
	return id == nilID
}

// String returns the custom base32 encoded representation of ID.
func (id ID) String() string {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	// avoids an allocation
	return *(*string)(unsafe.Pointer(&text))
}

// encode as Base32 using our custom character set
func encode(dst, id []byte) {
	encoding.Encode(dst, id[:])
}

// Bytes returns by value the byte array representation of ID.
func (id ID) Bytes() []byte {
	return id[:]
}

// Milliseconds returns the internal ID timestamp as the number of
// milliseconds from the Unix epoch.
func (id ID) Milliseconds() uint64 {
	return uint64(id[5]) |
		uint64(id[4])<<8 |
		uint64(id[3])<<16 |
		uint64(id[2])<<24 |
		uint64(id[1])<<32 |
		uint64(id[0])<<40
}

// Time returns the embedded timestamp value (milliseconds from the Unix epoch) as a
// time.Time value with millisecond resolution.
func (id ID) Time() time.Time {
	ms := id.Milliseconds()
	s := int64(ms / 1e3)
	ns := int64((ms % 1e3) * 1e6)
	return time.Unix(s, ns)
}

// Count returns the counter component of the ID.
func (id ID) Count() uint32 {
	return uint32(id[6])<<8 | uint32(id[7])
}

// FromString returns an ID by decoding a base32 representation of an ID
func FromString(str string) (ID, error) {
	id := &ID{}
	err := id.UnmarshalText([]byte(str))
	return *id, err
}

// FromBytes copies []bytes into an ID value.
func FromBytes(b []byte) (ID, error) {
	var id ID
	if len(b) != rawLen {
		return nilID, ErrInvalidID
	}
	copy(id[:], b)
	return id, nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// https://golang.org/pkg/encoding/#TextUnmarshaler
// All decoding is called from here.
func (id *ID) UnmarshalText(text []byte) error {
	// check for invalid length or characters in encoded id
	if len(text) != encodedLen {
		return ErrInvalidID
	}
	for _, c := range text {
		if dec[c] == 0xFF {
			return ErrInvalidID
		}
	}
	// buf := make([]byte, rawLen)
	// count, err := decode(buf, text)
	count, err := decode(id, text)
	if (count != rawLen) || (err != nil) {
		return ErrInvalidID
	}
	// copy(id[:], buf)
	return nil
}

// decode a Base32 representation of an ID as a []byte value.
func decode(id *ID, src []byte) (int, error) {
	return encoding.Decode(id[:], src)
}

// MarshalText implements encoding.TextMarshaler.
// https://golang.org/pkg/encoding/#TextMarshaler
func (id ID) MarshalText() ([]byte, error) {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return text, nil
}

// Value implements package sql's driver.Valuer.
// https://golang.org/pkg/database/sql/driver/#Valuer
func (id ID) Value() (driver.Value, error) {
	if id == nilID {
		return nil, nil
	}
	b, err := id.MarshalText()
	return string(b), err
}

// Scan implements the sql.Scanner interface.
// https://golang.org/pkg/database/sql/#Scanner
func (id *ID) Scan(value interface{}) (err error) {
	switch val := value.(type) {
	case string:
		return id.UnmarshalText([]byte(val))
	case []byte:
		return id.UnmarshalText(val)
	case nil:
		*id = nilID
		return nil
	default:
		return fmt.Errorf("sid: unsupported type: %T, value: %#v", value, value)
	}
}

// MarshalJSON implements the json.Marshaler interface.
// https://golang.org/pkg/encoding/json/#Marshaler
func (id ID) MarshalJSON() ([]byte, error) {
	// endless loop if merely return json.Marshal(id)
	if id == nilID {
		return []byte("null"), nil
	}
	text := make([]byte, encodedLen+2)
	encode(text[1:encodedLen+1], id[:])
	text[0], text[encodedLen+1] = '"', '"'
	return text, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// https://golang.org/pkg/encoding/json/#Unmarshaler
func (id *ID) UnmarshalJSON(text []byte) error {
	str := string(text)
	if str == "null" {
		*id = nilID
		return nil
	}
	return id.UnmarshalText(text[1 : len(text)-1])
}

// randInt generates a random number to initialize the counter. Despite the
// return value in the function signature, done for compatibility with package
// atomic functions, the actual value is deliberately constrained to uint16
// min/max values.
func randInt() uint32 {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("sid: cannot generate random number: %v;", err))
	}
	// casting to uint32 so we can utilize atomic.AddUint32 in NewWithTime().
	// Alternative to binary.BigEndian.Uint16(b)
	return uint32(uint16(b[0])<<8 | uint16(b[1]))
}
