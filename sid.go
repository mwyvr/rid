/*
Package sid provides a unique ID generator producing URL and human-friendly
(readability and double-click), compact, IDs. They are unique to a single
process, with more than 4 billion possibilities per millisecond.

The String() method produces a custom version of Base32 encoded IDs that look
like:

	af87cfy46ajbxf40 (16 characters, chronologically sortable)

The base32 encoding utilizes a customized alphabet based upon that popularized
by Crockford who replaced the more easily misread (by humans) i, o, l, and u
with the more easily read w, x, y, z. In sid, digits have been moved to the tail
of the character set to avoid having a leading zero for a great many years.

Each ID's 10-byte binary representation is comprised of a:

	001 125 209 022 154 224 016 025 151 086
	6-byte timestamp value representing milliseconds since the Unix epoch
	4-byte concurrency-safe counter (test included); maxCounter = uint32(4294967295)

The counter is initialized at a random value at initialization.

ID implements a number of common interfaces including package sql's
driver.Valuer, sql.Scanner, TextMarshaler, TextUnmarshaler, json.Marshaler,
json.Unmarshaler, and  Stringer.

Original source of inspiration:
https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

Acknowledgement: Much of this package is based on the globally-unique capable
rs/xid package which itself levers ideas from mongodb. See
https://github.com/rs/xid. I'd use xid if I had a fleet of apps on machines
spread around the world working in unison on a common datastore.

Comparisons:
	github.com/solutionroute/sid/v3:    af87cfy46ajbxf40
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
	rawLen     = 10                 // bytes
	encodedLen = 16                 // of base32 representation
	maxCounter = uint32(4294967295) // 4.29 billion IDs per millisecond

	//  ID string representations are 13 character long, Base32-encoded using a
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
	// counter is go routine safe and atomically updated. Initialized at a
	// random value between 0 and maxCounter (uint32 max: 4294967295), it's
	// protected from rollover back to 0, too.
	counter uint32

	// ErrInvalidID returned on attempt to decode an invalid ID character
	// representation (length or character set).
	ErrInvalidID = errors.New("sid: invalid id")

	// ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	nilID ID

	// dec is the decoding map for base32 encoding
	dec [256]byte
)

func init() {
	// We don't need crypto/rand
	rand.Seed(time.Now().UnixNano())
	counter = rand.Uint32()

	// create the base32 decoding table from the package charset
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
	// 4 byte counter, initialized at a random value.
	// Note: This implies max uint32 (4294967295) rollover is a possibility;
	// this is anticipated and is not an issue due to 4+ billion possibilities
	// per millisecond. Generating an ID is ~45ns on my current hardware;
	// therefore that limitation will never be reached.
	//
	// These operations are concurrency safe.
	atomic.CompareAndSwapUint32(&counter, maxCounter, 0)
	ct := atomic.AddUint32(&counter, 1)
	id[6] = byte(ct >> 24)
	id[7] = byte(ct >> 16)
	id[8] = byte(ct >> 8)
	id[9] = byte(ct)

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

// encode the fixed length id by unrolling the stdlib base32 algorithm +
// removing all safe checks; provides ~ 2x speed up over encoding/base32
// code adapted from github.com/rs/xid
func encode(dst, id []byte) {
	_ = dst[15] // eliminate compiler bounds checking
	_ = id[9]

	dst[15] = charset[id[9]&0x1F]
	dst[14] = charset[(id[9]>>5)|(id[8]<<3)&0x1F]
	dst[13] = charset[(id[8]>>2)&0x1F]
	dst[12] = charset[id[8]>>7|(id[7]<<1)&0x1F]
	dst[11] = charset[(id[7]>>4)&0x1F|(id[6]<<4)&0x1F]
	dst[10] = charset[(id[6]>>1)&0x1F]
	dst[9] = charset[(id[6]>>6)&0x1F|(id[5]<<2)&0x1F]
	dst[8] = charset[id[5]>>3]
	dst[7] = charset[id[4]&0x1F]
	dst[6] = charset[id[4]>>5|(id[3]<<3)&0x1F]
	dst[5] = charset[(id[3]>>2)&0x1F]
	dst[4] = charset[id[3]>>7|(id[2]<<1)&0x1F]
	dst[3] = charset[(id[2]>>4)&0x1F|(id[1]<<4)&0x1F]
	dst[2] = charset[(id[1]>>1)&0x1F]
	dst[1] = charset[(id[1]>>6)&0x1F|(id[0]<<2)&0x1F]
	dst[0] = charset[id[0]>>3]
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

// Count returns the counter value contained in the 4-byte count component of the ID.
func (id ID) Count() uint32 {
	return uint32(id[6])<<24 | uint32(id[7])<<16 | uint32(id[8])<<8 | uint32(id[9])
}

// FromString decodes a Base32 representation of an ID
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
	if len(text) != encodedLen {
		return ErrInvalidID
	}
	// check for invalid characters in encoded id supplied
	if len(text) != encodedLen {
		return ErrInvalidID
	}
	for _, c := range text {
		if dec[c] == 0xFF {
			return ErrInvalidID
		}
	}
	decode(id, text)
	return nil
}

// decode by unrolling the stdlib base32 algorithm + removing all safe checks
// code adapted from github.com/rs/xid
func decode(id *ID, src []byte) {
	_ = src[15] // eliminate compiler bounds checking
	_ = id[9]

	id[9] = dec[src[14]]<<5 | dec[src[15]]
	id[8] = dec[src[12]]<<7 | dec[src[13]]<<2 | dec[src[14]]>>3
	id[7] = dec[src[11]]<<4 | dec[src[12]]>>1
	id[6] = dec[src[9]]<<6 | dec[src[10]]<<1 | dec[src[11]]>>4
	id[5] = dec[src[8]]<<3 | dec[src[9]]>>2
	id[4] = dec[src[6]]<<5 | dec[src[7]]
	id[3] = dec[src[4]]<<7 | dec[src[5]]<<2 | dec[src[6]]>>3
	id[2] = dec[src[3]]<<4 | dec[src[4]]>>1
	id[1] = dec[src[1]]<<6 | dec[src[2]]<<1 | dec[src[3]]>>4
	id[0] = dec[src[0]]<<3 | dec[src[1]]>>2
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
