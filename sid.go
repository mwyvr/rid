/*
Package sid provides a ready-to-use generator of locally-unique IDs allowing
for more than 65 million IDs per second without collision.

IDs are [8]byte types composed of:

- 6-byte timestamp with millisecond resolution
- 2-byte counter (65536 per millisecond)

ID character representations are k-sortable Based32 encoded using a variant of
Crockford's alphabet. An example:

	af1zwtepacw38

Using sid is simple:

	package main

	import (
		"fmt"
		"github.com/solutionroute/sid"
	)

	func main(){
		id := sid.New()
		fmt.Printf("ID: %s Timestamp (ms): %d Count: %5d \nBytes: %3v\n",
			id.String(), id.Milliseconds(), id.Count(), id[:])
	}
	// ID: af3fwdh337xx6 Timestamp (ms): 1590631922127 Count: 26430
	// Bytes: [  1 114  89  12 249 207 103  62]

Using FromString("af1zwtepacw38") returns the original ID value:

	[  1 111  89  64 140   0 165 159] af1zwtepacw38 1577750400000 2019-12-30 16:00:00 -0800 PST counter: 42399

Acknowledgement: Much of this package was based off of the more capable rs/xid
package, which itself levers ideas from mongodb. See https://github.com/rs/xid.
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

// ID represents a locally unique identifier
type ID [rawLen]byte

const (
	rawLen     = 8  // bytes
	encodedLen = 13 // Base32

	// ID string representations are Base32-encoded using the Crockford
	// character set: i, o, l, u were removed and w, x, y, z added. To
	// avoid leading zeros for many years, the digits were moved last.
	// Standard for comparison: "0123456789abcdefghijklmnopqrstuv".
	charset = "abcdefghjkmnpqrstvwxyz0123456789" // mod-crockford
)

var (
	// base32 using the mod-crockford charset
	encoding = base32.NewEncoding(charset).WithPadding(-1)

	// counter is atomically updated and go routine-safe. While the type
	// is uint32, the value actually packed into ID is uint16 with a maximum
	// value of 65535; when hit it will loop to zero. This implies a maximum
	// of 65536 unique IDs per milliscond or 65,536,000 per second. Enough?
	counter = randInt()

	// ErrInvalidID is returned when trying to decode an invalid ID character
	// representation. See FromString and UnmarshalText.
	ErrInvalidID = errors.New("sid: invalid ID")

	nilID ID
)

// New returns a new ID using the current time.
func New() ID {
	return NewWithTime(time.Now())
}

// NewWithTime returns a new ID using the supplied Time value.
func NewWithTime(tm time.Time) ID {
	var id ID
	var ms uint64 // timestamp truncated to milliseconds

	ms = uint64(tm.Unix())*1000 + uint64(tm.Nanosecond()/int(time.Millisecond))
	// 6 bytes of time, to millisecond
	id[0] = byte(ms >> 40)
	id[1] = byte(ms >> 32)
	id[2] = byte(ms >> 24)
	id[3] = byte(ms >> 16)
	id[4] = byte(ms >> 8)
	id[5] = byte(ms)
	// 2 byte counter - rolls over at uint16 max
	// [1 114 88 144 14 181 255 255] af3fvear01998 1590623735477 65535  +1 =
	// [1 114 88 144 14 181   0   0] af3fvear0yaaa 1590623735477 0
	atomic.CompareAndSwapUint32(&counter, 65535, 0)
	ct := atomic.AddUint32(&counter, 1)
	id[6] = byte(ct >> 8)
	id[7] = byte(ct)

	return id
}

// String returns a Base32 representation of ID.
func (id ID) String() string {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return *(*string)(unsafe.Pointer(&text))
}

// Bytes returns by value the byte array representation of ID
func (id ID) Bytes() []byte {
	return id[:]
}

// Time returns the timestamp component as a Go time value with millisecond
// resolution.
func (id ID) Time() time.Time {
	ms := id.Milliseconds()
	s := int64(ms / 1e3)
	ns := int64((ms % 1e3) * 1e6)
	return time.Unix(s, ns)
}

// Milliseconds returns the timestamp of the ID as the number of milliseconds
// from the Unix epoch.
//
// Use the value from the ID.Time() method to access standard Unix()
// or UnixNano() timestamps.
func (id ID) Milliseconds() uint64 {
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

// FromString decodes a Base32 representation.
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

// encode ID as Base32
func encode(dst, id []byte) {
	encoding.Encode(dst, id[:])
}

// decode Base32 representation
func decode(buf []byte, src []byte) (int, error) {
	return encoding.Decode(buf, src)
}

// randInt generates a random number to initialize counter.
func randInt() uint32 {
	b := make([]byte, 2)
	if _, err := rand.Reader.Read(b); err != nil {
		panic(fmt.Errorf("sid: cannot generate random number: %v;", err))
	}
	// casting to uint32 so we can utilize atomic.AddUint32 in NewWithTime
	return uint32(uint16(b[0])<<8 | uint16(b[1]))
}

// Implementing interfaces for Text + SQL
// https://golang.org/src/encoding/encoding.go
// https://golang.org/src/database/sql/driver/types.go
// TODO https://golang.org/src/encoding/json/

// UnmarshalText implements encoding.TextUnmarshaler
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

// MarshalText implements encoding.TextMarshaler
func (id ID) MarshalText() ([]byte, error) {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return text, nil
}

// Value implements the driver.Valuer interface.
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
