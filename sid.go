/*
Package sid provides a no-configuration required unique ID generator for
applications with modest needs.

sid's 8-byte value can be stored directly as a 64 bit integer;both the byte
value and string representation are k-sortable.

sid produces URL and human-friendly (readability and double-click), compact IDs
(64 bit integer) with base32 encoded string (13 character) representations that
look like:

    af1zwtepacw38

Each ID's 8-byte binary representation:

    001 125 209 022 154 224 151 086

is comprised of a:

    6-byte timestamp value representing milliseconds since the Unix epoch
    2-byte concurrency-safe counter (test included); maxCounter = uint16(65535)

The counter is initialized at a random value at initialization.

*Modest Needs*

sid is intended for single process, single machine apps - perhaps using Go
friendly datastores like BoltDB, Badger or abstractions on top of either like
Genji.

The 2-byte concurrency-safe counter is a uint16, meaning 65,535 unique IDs can
be produced per millisecond or 1 ID every 16 nanoseconds. On the author's
hardware it takes more than 50ns to produce an ID, another 50ns to encode it,
longer yet to shove data into a datastore, so there's little chance of collision
- concurrency tests show this.

The base32 encoding utilizes a customized alphabet based upon that popularized
by Crockford who replaced the more easily misread (by humans) i, o, l, and u
with the more easily read w, x, y, z. In sid, digits have been moved to the tail
of the character set to avoid having a leading zero for a great many years.

ID implements a number of common interfaces including package sql's
driver.Valuer, sql.Scanner, TextMarshaler, TextUnmarshaler, json.Marshaler,
json.Unmarshaler, and  Stringer.

Original source of inspiration:
https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

Acknowledgement: Much of this package is based on the globally-unique capable
rs/xid package which itself levers ideas from mongodb. See
https://github.com/rs/xid. I'd use xid if I had a fleet of apps on machines
spread around the world working in unison on a common datastore.

Comparisons: github.com/solutionroute/sid/v3:    af87cfy46ajbxf40
    github.com/rs/xid:                  9bsv0s091sd002o20hk0
    github.com/segmentio/ksuid:         ZJkWubTm3ZsHZNs7FGt6oFvVVnD
    github.com/kjk/betterguid:          -HIVJnL-rmRZno06mvcV
    github.com/oklog/ulid:              014KG56DC01GG4TEB01ZEX7WFJ
    github.com/chilts/sid:              1257894000000000000-4601851300195147788
    github.com/lithammer/shortuuid:     DWaocVZPEBQB5BRMv6FUsZ
    github.com/google/uuid:             fa931eb3-cdc7-46a1-ae94-eb1b523203be

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

// ID represents a locally unique identifier
type ID [rawLen]byte

const (
	rawLen     = 8             // bytes
	encodedLen = 13            // of base32 representation
	maxCounter = uint32(65535) // IDs per millisecond

	//  ID string representations are Base32-encoded using a
	// character set proposed by *Crockford, but with digits following to avoid
	// producing IDs with leading zeros for many years.
	//
	//  *Crockford character set (i, o, l, u were removed and w, x, y, z added).
	//
	// encoding/Base32 standard for comparison:
	//        "0123456789abcdefghijklmnopqrstuv"
	charset = "abcdefghjkmnpqrstvwxyz0123456789"
)

var (
	// counter is atomically updated. Initialized at a random value between 0
	// and maxCounter, and is designed to rollover back to 0 (1, actually), too.

	counter uint32 // uint32 to take advantage of atomic pkg

	// ErrInvalidID returned on attempt to decode an invalid ID character
	// representation (length or character set).
	ErrInvalidID = errors.New("sid: invalid id")

	// ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	nilID ID

	// dec is the decoding map for base32 encoding
	dec      [256]byte
	encoding = base32.NewEncoding(charset).WithPadding(-1)
)

func init() {
	// We don't need crypto/rand
	rand.Seed(time.Now().UnixNano())
	counter = randInt()

	// create the base32 decoding table (used for error checking) from the charset
	for i := 0; i < len(dec); i++ {
		dec[i] = 0xFF
	}
	for i := 0; i < len(charset); i++ {
		dec[charset[i]] = byte(i)
	}
}

// New returns a new ID using the current time. ID's stored as []byte or as an
// encoded string have the same time resolution: milliseconds from the Unix
// epoch.
func New() ID {
	return NewWithTime(time.Now())
}

// NewWithTime returns a new ID based upon the supplied Time value.
func NewWithTime(tm time.Time) ID {
	var id ID
	// timestamp truncated to milliseconds
	var ms = uint64(tm.Unix())*1000 + uint64(tm.Nanosecond()/int(time.Millisecond))
	// id is 10 bytes:
	// 6 bytes of time, to millisecond
	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)
	// 2 byte counter, initialized at a random value.
	// These operations are concurrency safe.
	atomic.CompareAndSwapUint32(&counter, maxCounter, 0)
	ct := atomic.AddUint32(&counter, 1)
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

// Scan implements sql.Scanner.
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

// MarshalJSON implements json.Marshaler.
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

// UnmarshalJSON implements json.Unmarshaler.
// https://golang.org/pkg/encoding/json/#Unmarshaler
func (id *ID) UnmarshalJSON(text []byte) error {
	str := string(text)
	if str == "null" {
		*id = nilID
		return nil
	}
	return id.UnmarshalText(text[1 : len(text)-1])
}

// randInt generates a random number to initialize the counter.
// Despite the return value in the function signature, the actual value is
// deliberately constrained to uint16 min/max values.
func randInt() uint32 {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("sid: cannot generate random number: %v;", err))
	}
	// casting to uint32 so we can utilize atomic.AddUint32 in NewWithTime().
	// Alternative to binary.BigEndian.Uint16(b)
	return uint32(uint16(b[0])<<8 | uint16(b[1]))
}
