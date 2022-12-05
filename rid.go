/*
Package rid provides a k-sortable configuration-free, unique ID generator.

Package rid provides a [k-sortable](https://en.wikipedia.org/wiki/K-sorted_sequence),
configuration-free, unique ID generator.  Binary IDs Base-32 encode as a
20-character URL-friendly representation like: `ce0e7egs24nkzkn6egfg`.

The 12-byte binary representation of an ID is comprised of a:

  - 4-byte timestamp value representing seconds since the Unix epoch
  - 2-byte process signature comprised of md5 hash of machiine ID+process ID
  - 6-byte random value

rid implements a number of well-known interfaces to make
interacting with json and databases more convenient.  The String()
representation of ID is Base32 encoded using a modified Crockford-inspired
alphabet.

Example:

	id := rid.New()
	fmt.Printf("%s", id) 			//  cdym59rs24a5g86efepg
	fmt.Printf("%s", id.String()) 	//  cdym59rs24a5g86efepg

Acknowledgement: This package borrows heavily from rs/xid
(https://github.com/rs/xid), a capable globally-unique ID package which itself
levers ideas from MongoDB (https://docs.mongodb.com/manual/reference/method/ObjectId/).

Where rid differs from xid is in the use of (admittedly slower) random number
generation as opposed to a trailing counter for the last 4 bytes of the ID.
*/
package rid

import (
	"bytes"
	"crypto/md5"
	"database/sql/driver"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
	"unsafe"
)

// ID represents a locally unique, random-enough yet chronologically sortable identifier
type ID [rawLen]byte

const (
	rawLen     = 12 // binary representation
	encodedLen = 20 // base32 representation
	// charset stores the character set for a custom Base32 charset
	// inspired by Crockford: i, l, o, u removed and w, x, y, z added.
	//
	// charset/Base32 charset for comparison:
	//         "0123456789abcdefghijklmnopqrstuv"
	charset = "0123456789abcdefghjkmnpqrstvwxyz"
)

var (
	// nilID represents the zero-value of an ID
	// ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	nilID ID

	// rtSignature is derived from two bytes of the md5 hash of the machine
	// identifier and process ID
	rtSignature = runtimeSignature()

	encoding = base32.NewEncoding(charset).WithPadding(-1)
	// dec is the decoding map for base32 encoding
	dec [256]byte

	ErrInvalidID = errors.New("rid: invalid id")
)

func init() {
	// initialize the base32 decoding table
	for i := 0; i < len(dec); i++ {
		dec[i] = 0xFF
	}
	for i := 0; i < len(charset); i++ {
		dec[charset[i]] = byte(i)
	}
}

// New returns a new ID using the current time;
func New() ID {
	return NewWithTimestamp(uint32(time.Now().Unix()))
}

// func fastrand() uint64 {
// 	return uint64(new(maphash.Hash).Sum64())
// }

// NewWithTime returns a new ID based upon the supplied Time value.
func NewWithTimestamp(ts uint32) ID {
	var id ID
	// var randBytes = make([]byte, 6)

	binary.BigEndian.PutUint32(id[:], ts)
	// processSignature, only the first 2 bytes of the md5 hash
	id[4] = rtSignature[0]
	id[5] = rtSignature[1]
	// via fastrand()
	// rv := fastrand()
	rv := randUint64()
	id[6] = byte(rv >> 40)
	id[7] = byte(rv >> 32)
	id[8] = byte(rv >> 24)
	id[9] = byte(rv >> 16)
	id[10] = byte(rv >> 8)
	id[11] = byte(rv)
	return id
}

// IsNil returns true if ID == nilID
func (id ID) IsNil() bool {
	return id == nilID
}

// Alias of IsNil
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

// encode as Base32 using our custom character set
func encode(dst, id []byte) {
	encoding.Encode(dst, id[:])
}

// Bytes returns by value the byte slice representation of ID.
func (id ID) Bytes() []byte {
	return id[:]
}

// Seconds returns the ID timestamp component as seconds since the Unix epoch.
func (id ID) Seconds() int64 {
	return int64(binary.BigEndian.Uint32(id[0:4]))
}

// Time returns the ID's timestamp compoent as a Time value with seconds resolution.
func (id ID) Time() time.Time {
	return time.Unix(id.Seconds(), 0)
}

// RuntimeSignature returns the 2-byte identifier
func (id ID) RuntimeSignature() []byte {
	return id[4:6]
}

// Random returns the trailing random number component of the ID.
func (id ID) Random() uint64 {
	b := id[6:12]
	return uint64(uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 |
		uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5]))
}

// FromString returns an ID by decoding a Base32 representation of an ID
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
	for _, c := range text {
		// invalid characters (not in encoding)
		if dec[c] == 0xFF {
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
		return fmt.Errorf("rid: unsupported type: %T, value: %#v", value, value)
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

// Compare returns an integer comparing two IDs, comparing only the first 8 bytes:
// - 4-byte timestamp
// - 2-byte machine ID
// - 2-byte process ID
// ... while ignoring the trailing:
// - 4-byte random value
// Otherwise, it behaves just like `bytes.Compar(b1[:], b2[:])`.
// The result will be 0 if two IDs are identical, -1 if current id is less than
// the other one, and 1 if current id is greater than the other.
func (id ID) Compare(other ID) int {
	return bytes.Compare(id[:9], other[:9])
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

// randUint32 produces psuedo random numbers using the go runtime function fastrand
// These are non-deterministic and use on x86 hardware acceleration. 10x faster and
// passes uniqueness tests.
//
//go:linkname randUint32 runtime.fastrand
func randUint32() uint32

func randUint64() uint64 {
	return uint64(randUint32())<<32 | uint64(randUint32())
}

// runtimeSignature returns a md5 hash of a combination of a machine ID and the current process ID.
// If this function fails to get the hostname, and the fallback
// fails, it will cause a runtime error.
func runtimeSignature() []byte {
	sig := make([]byte, 2)
	hid, err := readPlatformMachineID()
	if err != nil || len(hid) == 0 {
		hid, err = os.Hostname()
	}
	if err != nil {
		// Fallback to rand number if machine id can't be gathered
		hid = strconv.Itoa(int(randUint32()))
	}
	pid := strconv.Itoa(os.Getpid())
	rs := md5.New()
	_, err = rs.Write([]byte(hid + pid)) // two strings
	if err != nil {
		panic(fmt.Errorf("rid: cannot produce signature hash: %v", err))
	}
	copy(sig, rs.Sum(nil))
	return sig
}
