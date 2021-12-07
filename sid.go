/*
Package sid provides a no-configuration required ID generator producing
compact, unique-enough (65535 per millisecond), URL and human-friendly IDs that
look like: af1zwtepacw38.

The 8-byte ID binary representation of ID is comprised of a:

	- 6-byte timestamp value representing milliseconds since the Unix epoch
	- 2-byte concurrency-safe counter (test included)

ID character representations (af1zwtepacw38) are 13 characters long,
chronologically-sortable and Base-32 encoded using a variant of Crockford's
alphabet.
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
	rawLen     = 8  // bytes
	encodedLen = 13 // Base32

	//  ID string representations are 13 character long, Base35-encoded using the
	//  Crockford character set (i, o, l, u were removed and w, x, y, z added).
	//  To avoid leading zeros for many years, the digits were moved last.
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
	// count is always in the range 1 - 65535
	atomic.CompareAndSwapUint32(&counter, 65535, 0)
	ct := atomic.AddUint32(&counter, 1)
	id[6] = byte(ct >> 8)
	id[7] = byte(ct)

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

// Time returns the embedded timestamp value as a time.Time value having
// millisecond resolution.
func (id ID) Time() time.Time {
	ms := id.Milliseconds()
	s := int64(ms / 1e3)
	ns := int64((ms % 1e3) * 1e6)
	return time.Unix(s, ns)
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

// Count returns the count component of the ID.
func (id ID) Count() uint16 {
	// Big-endian 2-byte value 0-65535
	return uint16(id[6])<<8 | uint16(id[7])
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
// deliberately constrained to uint16 min/max values.
func randInt() uint32 {
	b := make([]byte, 2)
	if _, err := rand.Reader.Read(b); err != nil {
		panic(fmt.Errorf("sid: cannot generate random number: %v;", err))
	}
	// casting to uint32 so we can utilize atomic.AddUint32 in NewWithTime
	return uint32(uint16(b[0])<<8 | uint16(b[1]))
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
