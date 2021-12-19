/*
Package sid provides a no-configuration required ID generator producing compact,
unique-enough (theoretically up to per millisecond), URL and human-friendly IDs that look
like: af1zwtepacw38.

The 8-byte ID binary representation of ID is comprised of a:

    - 6-byte timestamp value representing milliseconds since the Unix epoch
    - 2-byte concurrency-safe counter (test included)

ID character representations (e.g. af1zwtepacw38) are 13 characters long,
chronologically-sortable Base-32 encoded using an alphabet without
i, o, l, u, replaced instead with more easily identified (by humans)
w, x, y, z.

Limits:
Generating the maximum number of IDs per millisecond maxes out at one ID per
every 15 nanoseconds.

    1 millisecond / 65,535 = 15.2590219 nanoseconds

Encoding the ID []byte values as Base32 on my AMD Ryzen 7 3800X 8-Core Processor
takes ~55 nanoseconds.

Source of inspiration:
https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html


Much of this package is based on the globally-unique capable rs/xid package,
which itself levers ideas from mongodb.


I was bored during the early days of the COVID-19 pandemic is my excuse for
caring for a shorter unique-enough ID. Use another package if your app is
scaling beyond one machine.
Comparisons:
github.com/solutionroute/sid/v2:	af85984wnnt65cva
sid - 10 byte 						p5dy4eadsvg11r2e
github.com/rs/xid: 					9bsv0s091sd002o20hk0
github.com/segmentio/ksuid: 		ZJkWubTm3ZsHZNs7FGt6oFvVVnD
github.com/kjk/betterguid: 			-HIVJnL-rmRZno06mvcV
github.com/oklog/ulid: 				014KG56DC01GG4TEB01ZEX7WFJ
github.com/chilts/sid: 				1257894000000000000-4601851300195147788
github.com/lithammer/shortuuid: 	DWaocVZPEBQB5BRMv6FUsZ
github.com/google/uuid: 			fa931eb3-cdc7-46a1-ae94-eb1b523203be

*/
package sid

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/base32"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"
)

// Acknowledgement: Much of this package is based on the globally-unique capable
// rs/xid package which itself levers ideas from mongodb. See https://github.com/rs/xid.

// ID represents a locally unique identifier having a compact string representation.
type ID [rawLen]byte

const (
	rawLen     = 10 // bytes
	encodedLen = 16 // Base32
	maxCounter = uint32(4294967295)

	//  ID string representations are 13 character long, Base32-encoded using a
	// character set proposed by *Crockford, but with digits following to avoid
	// producing IDs with leading zeros for many years.
	//
	//  *Crockford character set (i, o, l, u were removed and w, x, y, z added).
	//
	// encoding/Base32 standard for comparison:
	//        "0123456789abcdefghijklmnopqrstuv".
	charset = "abcdefghjkmnpqrstvwxyz0123456789"
)

var (
	encoding = base32.NewEncoding(charset).WithPadding(-1)

	// counter is atomically updated and go-routine safe. While the type
	// is uint32, the value actually packed into ID is uint16 with a minimum
	// value of 1, maximum value of 65535; when max is hit, counter is reset.
	// This implies a maximum of 65535 unique IDs per millisecond or 65,535,000
	// per second.
	counter = randInt()

	// ErrInvalidID returned on attempt to decode an invalid ID character representation.
	ErrInvalidID = errors.New("sid: invalid ID")

	// ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	nilID ID
)

// New returns a new ID using the current time.
func New() ID {
	return NewWithTime(time.Now())
}

// NewWithTime returns a new ID based upon the supplied Time value.
func NewWithTime(tm time.Time) ID {
	var id ID
	// timestamp truncated to milliseconds
	var ms = uint64(tm.Unix())*1000 + uint64(tm.Nanosecond()/int(time.Millisecond))
	// 6 bytes of time, to millisecond
	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)
	// 4 byte counter - rolls over at uint32 max 4294967295
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

// String returns the Base32 encoded representation of ID.
func (id ID) String() string {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return *(*string)(unsafe.Pointer(&text))
}

// Bytes returns by value the byte array representation of ID.
func (id ID) Bytes() []byte {
	return id[:]
}

// Milliseconds returns the internal ID timestamp as the number of
// milliseconds from the Unix epoch.
//
// Use ID.Time() method to access standard Unix or UnixNano timestamps.
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

// Count returns the count component of the ID.
func (id ID) Count() uint32 {
	return uint32(id[6])<<24 | uint32(id[7])<<16 | uint32(id[8])<<8 | uint32(id[9])
}

// FromString decodes a Base32 representation to produce an ID.
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

// encode an ID as unpadded Base32 using the package encoding character set.
func encode(dst, id []byte) {
	encoding.Encode(dst, id[:])
}

// decode a Base32 representation of an ID as a []byte value.
func decode(buf []byte, src []byte) (int, error) {
	return encoding.Decode(buf, src)
}

// randInt generates a random number to initialize the counter.
// Despite the return value in the function signature, the actual value is
// deliberately constrained to uint16 values.
func randInt() uint32 {
	buf := make([]byte, 4)
	if _, err := rand.Reader.Read(buf); err != nil {
		panic(fmt.Errorf("sid: cannot generate random number: %v;", err))
	}
	// casting to uint32 so we can utilize atomic.AddUint32 in NewWithTime
	// return uint32(uint16(b[0])<<8 | uint16(b[1])<<24| uint16(b[0])<<8 | uint16(b[1]))

	return uint32(buf[0])<<24 | uint32(buf[1])<<16 | uint32(buf[2])<<8 | uint32(buf[3])
}

// UnmarshalText implements encoding.TextUnmarshaler.
// https://golang.org/pkg/encoding/#TextUnmarshaler
func (id *ID) UnmarshalText(text []byte) error {
	if len(text) != encodedLen {
		return ErrInvalidID
	}
	buf := make([]byte, rawLen)
	count, err := decode(buf, text)
	if (count != rawLen) || (err != nil) {
		return ErrInvalidID
	}
	copy(id[:], buf)
	return nil
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

// MarshalJSON implements json.Masrshaler.
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
