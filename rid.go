/*
Package rid provides a performant, k-sortable, scalable unique ID generator
suitable for applications where ID generation coordination between machines or
other processes is not required. ID generation is goroutine safe and scales
well with CPU cores. Providing unique non-sequential keys for embeddable
databases like SQLIte or BoltDB or key-value stores are typical use-cases.

Binary IDs Base-32 encode as a 16-character URL and human-friendly
representation like dfp7qt0v2pwt0v2x.

The 10-byte binary representation of an ID is comprised of:

  - 4-byte timestamp value representing seconds since the Unix epoch
  - 6-byte random value; see fastrand48 [1]

Key features:

  - K-orderable in both binary and string representations
  - Encoded IDs are short (16 characters)
  - Automatic (de)serialization for SQL and JSON
  - Scalable performance as cores increase; ID generation is fast and remains so
  - URL and human friendly Base32 encoding using a custom character set to
    avoid unintended rude words if humans are to be exposed to IDs

Example usage:

	id := rid.New()
	fmt.Printf("%s", id.String())
	// Output: dfp7qt97menfv8ll
	id, err := id.FromString("dfp7qt97menfv8ll")
	// ID{0x63,0xac,0x7b,0xe9,0x27,0xa3,0x6a,0xed,0xa2,0x73}, nil

Acknowledgement: This source file is based on work in package github.com/rs/xid,
a zero-configuration globally-unique ID generator. See LICENSE.rs-xid.
The same API has been maintained.
*/
package rid

import (
	"bytes"
	"database/sql/driver"
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
	rawLen     = 10                                 // binary
	encodedLen = 16                                 // base32
	charset    = "0123456789bcdefghkjlmnpqrstvwxyz" // fewer vowels to avoid random rudeness
)

var (
	// nilID represents the zero-value of an ID
	nilID ID

	// dec provides a decoding map
	dec [256]byte

	// ErrInvalidID represents errors in converting from invalid
	// []byte or string representations
	ErrInvalidID = errors.New("rid: invalid id")
)

func init() {
	// initialize the decoding map, used also for sanity checking input
	for i := 0; i < len(dec); i++ {
		dec[i] = 0xFF
	}
	for i := 0; i < len(charset); i++ {
		dec[charset[i]] = byte(i)
	}
}

// New returns a new ID using the current time.
func New() ID {
	return NewWithTime(time.Now())
}

// NewWithTime returns a new ID using the supplied time.
//
// The time value component of an ID is a Unix timestamp with seconds
// resolution.
func NewWithTime(t time.Time) ID {
	var id ID

	// 4 bytes of time
	binary.BigEndian.PutUint32(id[:], uint32(t.Unix()))
	// 6 bytes of randomness
	r := fastrand48()
	id[4] = byte(r >> 40)
	id[5] = byte(r >> 32)
	id[6] = byte(r >> 24)
	id[7] = byte(r >> 16)
	id[8] = byte(r >> 8)
	id[9] = byte(r)
	return id
}

// IsNil returns true if ID == nilID.
func (id ID) IsNil() bool {
	return id == nilID
}

// IsZero is an alias of is IsNil.
func (id ID) IsZero() bool {
	return id.IsNil()
}

// NilID returns a zero value for `rid.ID`.
func NilID() ID {
	return nilID
}

// String returns id as Base32 encoded string.
func (id ID) String() string {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return *(*string)(unsafe.Pointer(&text))
}

// Encode id, writing 16 bytes to dst and returning it.
func (id ID) Encode(dst []byte) []byte {
	encode(dst, id[:])
	return dst
}

// encode bytes as Base32, unrolling the stdlib base32 algorithm for
// performance. There is no padding as Base32 aligns on 5-byte boundaries.
func encode(dst, id []byte) {
	// minor bounds checking compiler optimization
	// go tool compile -d=ssa/check_bce/debug=1 rid.go
	_ = id[9]
	_ = dst[15]

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

// Bytes returns the binary representation of ID.
func (id ID) Bytes() []byte {
	return id[:]
}

// Timestamp returns the ID's timestamp component as seconds since the Unix epoch.
func (id ID) Timestamp() int64 {
	b := id[0:4]
	// Big Endian
	return int64(uint64(b[0])<<24 | uint64(b[1])<<16 | uint64(b[2])<<8 | uint64(b[3]))
}

// Time returns the ID's timestamp as a Time value.
func (id ID) Time() time.Time {
	return time.Unix(id.Timestamp(), 0)
}

// Random returns the random component of the ID as an unsigned integer.
func (id ID) Random() uint64 {
	b := id[4:]
	// Big Endian
	return uint64(uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 | uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5]))
}

// FromString decodes a Base32 encoded string to return an ID.
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

// UnmarshalText implements encoding.TextUnmarshaler
// https://golang.org/pkg/encoding/#TextUnmarshaler
// All decoding is called from here.
func (id *ID) UnmarshalText(text []byte) error {
	if len(text) != encodedLen {
		*id = nilID
		return ErrInvalidID
	}
	// characters not in the decoding map will return an error
	for _, c := range text {
		if dec[c] == 0xFF {
			return ErrInvalidID
		}
	}
	if !decode(id, text) {
		*id = nilID
		return ErrInvalidID
	}
	return nil
}

// decode a Base32 encoded string by unrolling the stdlib Base32 algorithm.
func decode(id *ID, src []byte) bool {
	// bounds checking compiler optimization
	// go tool compile -d=ssa/check_bce/debug=1 rid.go
	_ = src[15]

	// this is ~4 to 6x faster than stdlib Base32 decoding
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
	return true
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
	if id.IsNil() {
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
		return fmt.Errorf("rid: scanning unsupported type: %T", value)
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
func (id *ID) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str == "null" {
		*id = nilID
		return nil
	}
	// Check the slice length to prevent runtime bounds check panic in UnmarshalText()
	if len(b) < 2 {
		return ErrInvalidID
	}
	return id.UnmarshalText(b[1 : len(b)-1])
}

// Compare makes IDs k-sortable, returning an integer comparing only the first
// 4 bytes of two IDs.
//
// Recall that an ID is comprized of a:
//
// - 4-byte timestamp
// - 6-byte random value
//
// Otherwise, it behaves just like `bytes.Compare(b1[:], b2[:])`.
//
// The result will be 0 if two IDs are identical, -1 if current id is less than
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

// Sort sorts an array of IDs in place.
func Sort(ids []ID) {
	sort.Sort(sorter(ids))
}

// [1] Random number generation: For performance and in particular scalability,
// this package uses an internal runtime Go function fastrand64.
//
// For proof of utility in concurrent situations see eval/uniqcheck/main.go.
//
// For more information on fastrand see the Go source at:
// https://cs.opensource.google/go/go/+/master:src/runtime/stubs.go?q=fastrand

// Link the runtime to a non-exported rid package name:
//
//go:linkname fastrand64 runtime.fastrand64
func fastrand64() uint64

const maxRandom uint64 = 0xffffffffffff // 2^48

// fastrand48 returns from a 64 bit source a random value capped at 48 bits (6
// bytes) using the same approach as the Go runtime fastrandn().
func fastrand48() uint64 {
	// See https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	return fastrand64() * maxRandom >> 16
}
