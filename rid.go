/*
Package rid provides a k-sortable unique ID generator suitable for situations
where inter-process coordination of ID generation is not required.

IDs encode as a 16-character URL-friendly representation like
dfp7qt0v2pwt0v2x.

The 10-byte binary representation of an rid.ID is comprised of:

  - 4-byte timestamp value representing seconds since the Unix epoch
  - 6-byte (48 bits) random value (per second)

Method String() base32 encodes ID using a custom character set missing
characters "a,i,o,u" to reduce unintentional random rudeness.
*/
package rid

import (
	"bytes"
	"database/sql/driver"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
	"time"
	"unsafe"
)

// ID represents a unique identifier
type ID [rawLen]byte

const (
	rawLen     = 10 // binary length in bytes
	encodedLen = 16 // string encoded length in bytes

	// A modifed "Crockford" approach, this 32 character encoding map has no
	// "a,i,o,u" to greatly reduce random rudeness for cases where humans see IDs
	charset = "0123456789bcdefghkjlmnpqrstvwxyz"
)

var (
	nilID ID

	// customized base32 encoder/decoder
	encoding = base32.NewEncoding(charset)

	ErrInvalidID = errors.New("rid: invalid id")
)

// New returns a new ID using the current time.
func New() ID {
	return NewWithTime(time.Now())
}

// NewWithTime returns a new ID using the supplied Time.
//
// Note: The time value component of an ID is a Unix timestamp; in Go,
// timestamp values are identical regardless of the Time value's location,
// i.e.: time.Now().Unix() == time.Now().UTC().Unix()
func NewWithTime(t time.Time) ID {
	var id ID

	// 4 bytes of time, representing seconds since Unix epoch
	binary.BigEndian.PutUint32(id[:], uint32(t.Unix()))
	// take an 8 byte random number and cap at max value for 6 bytes
	r := fastrandUint64() * 0xffffffffffff >> 16 // equiv but slightly faster than fastrandUint64() % 0xffffffffffff
	id[4] = byte(r >> 40)
	id[5] = byte(r >> 32)
	id[6] = byte(r >> 24)
	id[7] = byte(r >> 16)
	id[8] = byte(r >> 8)
	id[9] = byte(r)
	return id
}

// FromString reads an ID from its string represetnation
func FromString(str string) (ID, error) {
	id := &ID{}
	err := id.UnmarshalText([]byte(str))
	return *id, err
}

// String returns ID as a string representation, encoded as base32 using a custom charset
func (id ID) String() string {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return *(*string)(unsafe.Pointer(&text))
}

// IsNil returns true if ID == nilID
func (id ID) IsNil() bool {
	return id == nilID
}

// IsZero is an alias of is IsNil
func (id ID) IsZero() bool {
	return id.IsNil()
}

// NilID returns a zero value for `rid.ID`.
func NilID() ID {
	return nilID
}

// Bytes returns by value the binary representation of ID.
func (id ID) Bytes() []byte {
	return id[:]
}

// Timestamp returns the ID's timestamp component as seconds since the Unix epoch.
func (id ID) Timestamp() int64 {
	return int64(binary.BigEndian.Uint32(id[0:4]))
}

// Time returns the ID's timestamp as a Time value
func (id ID) Time() time.Time {
	return time.Unix(id.Timestamp(), 0)
}

// Random returns the random number component of the ID
func (id ID) Random() uint64 {
	b := id[4:]
	// Random is stored as a six-byte big endian value
	return uint64(uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 | uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5]))
}

// FromBytes copies []bytes into an ID value
func FromBytes(b []byte) (ID, error) {
	var id ID
	if len(b) != rawLen {
		return nilID, ErrInvalidID
	}
	copy(id[:], b)
	return id, nil
}

// UnmarshalText implements encoding.TextUnmarshaler
// https://golang.org/pkg/encoding/#TextUnmarshaler
// All decoding is called from here.
func (id *ID) UnmarshalText(text []byte) error {
	if len(text) != encodedLen {
		*id = nilID
		return ErrInvalidID
	}
	// invalid characters will return an error
	if _, err := decode(id, text); err != nil {
		*id = nilID
		return ErrInvalidID
	}
	return nil
}

// decode a Base32 encoded ID
func decode(id *ID, src []byte) (int, error) {
	return encoding.Decode(id[:], src)
}

// MarshalText implements encoding.TextMarshaler
// https://golang.org/pkg/encoding/#TextMarshaler
func (id ID) MarshalText() ([]byte, error) {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return text, nil
}

// encode id using the customized base32 encoding
func encode(dst, id []byte) {
	encoding.Encode(dst, id[:])
}

// Value implements package sql's driver.Valuer
// https://golang.org/pkg/database/sql/driver/#Valuer
func (id ID) Value() (driver.Value, error) {
	if id.IsNil() {
		return nil, nil
	}
	b, err := id.MarshalText()
	return string(b), err
}

// Scan implements the sql.Scanner interface
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
		return fmt.Errorf("rid: scanning unsupported type: %T", value)
	}
}

// MarshalJSON implements the json.Marshaler interface
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

// UnmarshalJSON implements the json.Unmarshaler interface
// https://golang.org/pkg/encoding/json/#Unmarshaler
func (id *ID) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str == "null" {
		*id = nilID
		return nil
	}
	// Check the slice length to prevent panic on passing it to UnmarshalText()
	if len(b) < 2 {
		return ErrInvalidID
	}
	return id.UnmarshalText(b[1 : len(b)-1])
}

// Compare makes IDs k-sortable, returning an integer comparing two IDs,
// comparing only the first 4 bytes:
//
//   - 4-byte timestamp
//   - 6-byte random value
//
// Otherwise, it behaves just like `bytes.Compare(b1[:], b2[:])`. The result
// will be 0 if two IDs are identical, -1 if current id is less than
// the other one, and 1 if current id is greater than the other.
func (id ID) Compare(other ID) int {
	return bytes.Compare(id[:5], other[:5])
}

type sorter []ID

func (s sorter) Len() int {
	return len(s)
}

func (s sorter) Less(i, j int) bool {
	return s[i].Compare(s[j]) < 0
}

func (s sorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Sort sorts an array of IDs inplace.
// It works by wrapping `[]ID` and use `sort.Sort`.
func Sort(ids []ID) {
	sort.Sort(sorter(ids))
}

// [1] Random number generation: For performance and scalability, this package
// uses an internal Go function `fastrand64`. See eval/uniqcheck/main.go for
// a proof of utility.
//
// For more information on fastrand see the Go source at:
// https://cs.opensource.google/go/go/+/master:src/runtime/stubs.go?q=fastrand
// which include the comments:
//
// 	Implement wyrand: https://github.com/wangyi-fudan/wyhash
// 	Implement xorshift64+: 2 32-bit xorshift sequences added together.
// 	Xorshift paper: https://www.jstatsoft.org/article/view/v008i14/xorshift.pdf
// 	This generator passes the SmallCrush suite, part of TestU01 framework:
// 	http://simul.iro.umontreal.ca/testu01/tu01.html
//
// Note: importing unexported functions requires a module import package
// unsafe; this package already does, so...

//go:linkname fastrandUint64 runtime.fastrand64
func fastrandUint64() uint64
