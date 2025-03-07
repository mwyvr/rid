/*
Package rid provides a performant goroutine-safe ID generator producing unique,
k-sortable, short, url-safe IDs suitable where ID generation coordination across
the globe is not a requirement.

The 10-byte binary representation of an ID is comprised of:

  - 6-byte timestamp value representing milliseconds since the Unix epoch.
  - 2-byte ordered sequence
  - 2-bytes of random data;  random value; as of release v1.2.0 this package
    uses crypto/rand and requires Go 1.24+.

The millisecond << 12 plus sequence value are guaranteed to
be greater than the previous call(s) to New().

IDs encode (base32) as 16 character human and url-friendly strings.

Key ID features:

  - K-orderable in both binary and string representations
  - Automatic (de)serialization for SQL and JSON
  - Encoded IDs are short (16 characters)
  - URL and human friendly Base32 encoding using a custom character set to
    avoid unintended rude words if humans are to be exposed to IDs

Example usage:

	id := rid.New()
	fmt.Printf("%s", id.String())
	// Output: 06bpw16hfm62jt9h
	id, err := id.FromString("06bpw16hfm62jt9h")
	// ID{  0x1, 0x95, 0x6e,  0x4, 0xd0, 0x75,  0xc, 0x29, 0x69, 0x30 }, nil

Acknowledgement: While the ID payload differs greatly, the API and much of
this package is based on on package github.com/rs/xid, a zero-configuration
globally-unique ID generator.
*/
package rid

import (
	"bytes"
	"crypto/rand"
	"database/sql/driver"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ID represents a unique identifier
type ID [rawLen]byte

const (
	rawLen     = 10                                 // binary
	encodedLen = 16                                 // base32
	charset    = "0123456789bcdefghkjlmnpqrstvwxyz" // fewer vowels to avoid random rudeness
	maxByte    = 0xFF                               // used as a sentinel value in charmap
)

var (
	// nilID represents the zero-value of an ID
	nilID ID

	// dec provides a decoding map
	dec [256]byte

	// ErrInvalidID represents errors returned when converting from invalid
	// []byte, string or json representations
	ErrInvalidID = errors.New("rid: invalid id")
)

func init() {
	// initialize the decoding map, used also for sanity checking input
	for i := range len(dec) {
		dec[i] = maxByte
	}
	for i := range len(charset) {
		dec[charset[i]] = byte(i)
	}
}

// New returns a new ID using the current time.
// The time component of an ID is a Unix timestamp in milliseconds resolution;
func New() ID {
	var id ID

	t, s := getTS() // (timestamp+sequence) are guaranteed to be unique for each call
	// time
	id[0] = byte(t >> 40)
	id[1] = byte(t >> 32)
	id[2] = byte(t >> 24)
	id[3] = byte(t >> 16)
	id[4] = byte(t >> 8)
	id[5] = byte(t)
	// sequence
	id[6] = byte(s >> 8)
	id[7] = byte(s)
	// two bytes of randomness
	rand.Read(id[8:])
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

// String returns id as Base32 encoded string using a Crockford character set.
func (id ID) String() string {
	text := make([]byte, encodedLen)
	encode(text, id[:])
	return string(text)
}

// Encode id, writing 16 bytes to dst and returning it.
func (id ID) Encode(dst []byte) []byte {
	encode(dst, id[:])
	return dst
}

// encode bytes as Base32, unrolling the stdlib base32 algorithm for
// performance. There is no padding as Base32 aligns on 5-byte boundaries.
func encode(dst, id []byte) {

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

// Timestamp returns the ID's timestamp component as milliseconds since the
// Unix epoch.
func (id ID) Timestamp() int64 {
	b := id[0:6]
	// Big Endian
	return int64(uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 | uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5]))
}

// Sequence returns the ID sequence.
func (id ID) Sequence() int64 {
	b := id[6:8]
	// Big Endian
	return int64(uint64(b[0])<<8 | uint64(b[1]))
}

// Time returns the ID's timestamp as a Time value.
func (id ID) Time() time.Time {
	return time.UnixMilli(id.Timestamp()).UTC()
}

// Random returns the random component of the ID.
func (id ID) Random() uint64 {
	b := id[8:]
	// Big Endian
	return uint64(b[0])<<8 | uint64(b[1])
}

// FromString decodes a Base32-encoded string to return an ID.
func FromString(str string) (ID, error) {
	id := &ID{}
	err := id.UnmarshalText([]byte(str))

	return *id, err
}

// FromBytes copies []bytes into an ID value. For validity, only a length-check
// is possible and performed.
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
		if dec[c] == maxByte {
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
	// this is ~4 to 6x faster than stdlib Base32 decoding
	id[9] = dec[src[14]]<<5 | dec[src[15]]
	// check the last byte
	if charset[id[9]&0x1F] != src[15] {
		return false
	}
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
func (id *ID) Scan(value any) (err error) {
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

	text := make([]byte, encodedLen+2) // 2 = len of ""
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

// Compare makes IDs k-sortable, returning an integer comparing only the
// first 6 bytes of two IDs.
//
// Recall that an ID is comprized of a:
//
// - 6-byte timestamp
// - 2-byte sequence
// - 2-byte random value
//
// Otherwise, it behaves just like `bytes.Compare(b1[:], b2[:])`.
//
// The result will be 0 if two IDs are identical, -1 if current id is less than
// the other one, and 1 if current id is greater than the other.
func (id ID) Compare(other ID) int {
	return bytes.Compare(id[:8], other[:8])
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

// getTS is borrowed directly from getV7Time:
// https://github.com/google/uuid/blob/2d3c2a9cc518326daf99a383f07c4d3c44317e4d/version7.go#L88

var (
	// lastTime is the last time we returned stored as:
	//
	//	52 bits of time in milliseconds since epoch
	//	12 bits of (fractional nanoseconds) >> 8
	lastTime int64
	timeMu   sync.Mutex
	timeNow  = time.Now // for testing
)

const nanoPerMilli = 1000000

// getTS using the supplied time func, returns the time in milliseconds and
// nanoseconds / 256.
//
// The returned (milli << 12 + seq) is guaranteed to be greater than
// (milli << 12 + seq) returned by any previous call to getTS.
func getTS() (milli, seq int64) {
	timeMu.Lock()
	defer timeMu.Unlock()

	nano := timeNow().UnixNano()
	// fmt.Printf("%v\n", tf())
	milli = nano / nanoPerMilli
	// Sequence number is between 0 and 3906 (nanoPerMilli>>8)
	seq = (nano - milli*nanoPerMilli) >> 8
	now := milli<<12 + seq
	if now <= lastTime {
		now = lastTime + 1
		milli = now >> 12
		seq = now & 0xfff
	}
	lastTime = now
	return milli, seq
}
