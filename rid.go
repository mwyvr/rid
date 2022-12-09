/*
Package rid provides a k-sortable configuration-free, unique ID generator.

Package rid provides a [k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence),
configuration-free, unique ID generator.  Binary IDs Base-32 encode as a
20-character URL-friendly representation like: `ce0e7egs24nkzkn6egfg`.

The 15-byte binary representation of an ID is comprised of a:

  - 6-byte timestamp value representing milliseconds since the Unix epoch
  - 1-byte machine+process signature, derived from a md5 hash of the machine ID + process ID
  - 8-byte random number provided by Go's runtime `fastrand64` function. [1]

Features:

  - Size: 15 bytes, larger than xid, smaller than UUID
  - Base32 encoded by default, case insensitive and URL-friendly
  - Base32 aligns to multiples of 5 bytes
  - K-orderable in both binary and string representations
  - Embedded time, having millisecond prescision
  - 1 byte representing machine+pid
  - 48 bits of randomness
  - scalable performance

rid also implements a number of well-known interfaces to make interacting with json
and databases more convenient.

Example:

	id := rid.New()
	fmt.Printf("%s", id.String()) 	// 062f00pexapx8rzhnyefnj2w

Acknowledgement: This package borrows heavily from rs/xid
(https://github.com/rs/xid), a capable globally-unique ID package which itself
levers ideas from MongoDB (https://docs.mongodb.com/manual/reference/method/ObjectId/).
*/
package rid

import (
	"bytes"
	"crypto/md5"
	"database/sql/driver"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
	"unsafe"
)

// ID represents a unique identifier
type ID [rawLen]byte

const (
	rawLen     = 15 // binary
	encodedLen = 24 // base32 representation
	// crockford stores the character set for a custom Base32 character set
	// inspired by Crockford: i, l, o, u removed and w, x, y, z added.
	//
	// crockford/Base32 crockford for comparison:
	//        "0123456789abcdefghijklmnopqrstuv"
	crockford = "0123456789abcdefghjkmnpqrstvwxyz"
)

var (
	// nilID represents the zero-value of an ID
	// ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	nilID ID

	// md5 hash of the machine identifier and process ID
	rtsig = runtimeSignature()

	// rid default encoding
	encoding = base32.NewEncoding(crockford).WithPadding(-1)
	// used by String64 and FromString64 helper functions
	encoding64 = base64.URLEncoding

	// dec is a decoding map, used only for Base32 validity checking
	dec [256]byte

	ErrInvalidID = errors.New("rid: invalid id")
)

func init() {
	// initialize the base32 decoding table
	for i := 0; i < len(dec); i++ {
		dec[i] = 0xFF
	}
	for i := 0; i < len(crockford); i++ {
		dec[crockford[i]] = byte(i)
	}
}

// New returns a new ID using the current time;
func New() ID {
	return NewWithTimestamp(uint64(time.Now().UnixMilli()))
}

// NewWithTimestamp returns a new ID using the supplied timestamp
func NewWithTimestamp(ts uint64) ID {
	var id ID

	// 6 byte timestamp (milliseconds since Unix epoch)
	id[0] = byte(ts >> 40)
	id[1] = byte(ts >> 32)
	id[2] = byte(ts >> 24)
	id[3] = byte(ts >> 16)
	id[4] = byte(ts >> 8)
	id[5] = byte(ts)
	// 1 byte runtime signature, semi-random
	id[6] = rtsig[0]
	// 8 bytes of randomnesss
	binary.BigEndian.PutUint64(id[7:], randUint64())
	return id
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

// String returns the custom base32 encoded representation of ID as a string.
func (id ID) String() string {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	// avoids an allocation
	return *(*string)(unsafe.Pointer(&text))
}

// encode bytes as Base32 using our custom character set
func encode(dst, id []byte) {
	encoding.Encode(dst, id[:])
}

// Bytes returns by value the binary representation of ID.
func (id ID) Bytes() []byte {
	return id[:]
}

// Timestamp returns the ID's timestamp component.
func (id ID) Timestamp() int64 {
	b := id[0:6]
	// Big Endian
	return int64(uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 | uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5]))
}

// Time returns the ID's timestamp as a Time value
func (id ID) Time() time.Time {
	return time.UnixMilli(id.Timestamp())
}

// RuntimeSignature returns the ID's machine id/pid signature
func (id ID) RuntimeSignature() []byte {
	return id[6:7]
}

// Random returns the random number component of the ID
func (id ID) Random() uint64 {
	return binary.BigEndian.Uint64(id[7:])
}

// FromString decodes the supplied Base32 representation
func FromString(str string) (ID, error) {
	id := &ID{}
	err := id.UnmarshalText([]byte(str))
	return *id, err
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
	for _, c := range text {
		// invalid characters (not in encoding)
		if dec[c] == 0xFF {
			*id = nilID
			return ErrInvalidID
		}
	}
	if _, err := decode(id, text); err != nil {
		*id = nilID
		return ErrInvalidID
	}
	return nil
}

// decode a Base32 representation of an ID
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
		return fmt.Errorf("rid: unsupported type: %T, value: %#v", value, value)
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
// comparing only the first 7 bytes:
//
//   - 6-byte timestamp
//   - 1-byte runtime signature
//     ... while ignoring the trailing:
//   - 6-byte random value
//
// Otherwise, it behaves just like `bytes.Compare(b1[:], b2[:])`. The result
// will be 0 if two IDs are identical, -1 if current id is less than
// the other one, and 1 if current id is greater than the other.
func (id ID) Compare(other ID) int {
	return bytes.Compare(id[:7], other[:7])
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

// Alternative Base64 encoding/decoding helpers
// String64 returns a Base64 encoded representation of ID as a string.
func String64(id ID) string {
	text := make([]byte, (rawLen/3)*4)
	encoding64.Encode(text, id[:])
	// avoids an allocation
	return *(*string)(unsafe.Pointer(&text))
}

// FromString64 returns an ID by decoding a Base64 representation of an ID
func FromString64(str string) (ID, error) {
	encoded64Len := (rawLen / 3) * 4 // 20
	id := ID{}
	if len(str) != encoded64Len {
		return id, ErrInvalidID
	}
	if _, err := encoding64.Decode(id[:], []byte(str)); err != nil {
		return id, ErrInvalidID
	} else {
		return id, err
	}
}

// runtimeSignature returns the first byte of a md5 hash of (machine ID + process ID).
// If this function fails it will cause a runtime error.
func runtimeSignature() []byte {
	sig := make([]byte, 1)
	hwid, err := readPlatformMachineID()
	if err != nil || len(hwid) == 0 {
		// fallback to hostname (common)
		hwid, err = os.Hostname()
	}
	if err != nil {
		// Fallback to rand number if both machine ID (common possibility) AND
		// hostname (uncommon possibility) can't be read.
		hwid = strconv.Itoa(int(randUint64()))
	}
	pid := strconv.Itoa(os.Getpid())
	rs := md5.New()
	_, err = rs.Write([]byte(hwid + pid))
	if err != nil {
		panic(fmt.Errorf("rid: cannot produce signature hash: %v", err))
	}
	copy(sig, rs.Sum(nil))
	return sig
}

// [1] Random number generation:
// For performance and scalability, this package uses an unexposed internal
// Go function `fastrand` / `fastrand64`. See eval/uniqcheck/main.go.
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

//go:linkname randUint64 runtime.fastrand64
func randUint64() uint64
