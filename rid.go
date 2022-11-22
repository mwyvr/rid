/*
Package rid provides a unique-enough ID generator for
applications with modest (meaning non-distributed) needs. 

rid can generate more than 4 billion possible random IDs per second, per
process, per machine. Duplication in practical use is unlikely. Ymmv.

    id := rid.New()
    fmt.Printf("%s", id) 
// TODO replace all the examples once fully tested

Each ID's 12-byte binary representation is comprised of a:

- 4-byte timestamp value representing seconds since the Unix epoch
- 2-byte machine ID
- 2-byte process ID
- 4-byte random value

IDs are chronologically sortable to the second, with a tradeoff in fine-grained
sortability due to the trailing random value component.

The String() representation is Base32 encoded using a modified Crockford
inspired alphabet.

Acknowledgement: Much of this package is based on the globally-unique capable
rs/xid package which itself levers ideas from mongodb. See https://github.com/rs/xid.
*/
package rid

import (
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
    "io/ioutil"
	"crypto/md5"
	"crypto/rand"
    "os"
	"time"
	"unsafe"
)

// ID represents a locally unique, random-enough yet chronologically sortable identifier
type ID [rawLen]byte

const (
	rawLen     = 12                  // binary representation
	encodedLen = 20                 // base32 representation
    // ID string representations are base32-encoded using a character set
    // inspired by Crockford: i, l, o, u removed and w, x, y, z added.
    // 
    // encoding/Base32 charset for comparison:
    //        "0123456789abcdefghijklmnopqrstuv"
	encoding = "0123456789abcdefghjkmnpqrstvwxyz"
)

var (
	// ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	nilID ID

	// machineId stores a md5 hash of the machine identifier or hostname
	machineID = readMachineID()

	// pid stores the current process id
	pid = os.Getpid()

	ErrInvalidID = errors.New("rid: invalid id")
	ErrInvalidLength = errors.New("rid: invalid encoded length")

	// dec is the decoding map for base32 encoding
	dec      [256]byte
)

func init() {
	// initialize the base32 decoding table
	for i := 0; i < len(dec); i++ {
		dec[i] = 0xFF
	}
	for i := 0; i < len(encoding); i++ {
		dec[encoding[i]] = byte(i)
	}
}

// New returns a new ID using the current time; IDs represent millisecond time resolution.
func New() ID {
	return NewWithTime(time.Now())
}

// NewWithTime returns a new ID based upon the supplied Time value.
func NewWithTime(tm time.Time) ID {
	var id ID

	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(id[:], uint32(tm.Unix()))
	// Machine, only the first 2 bytes of md5(hostname)
	id[4] = machineID[0]
	id[5] = machineID[1]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	id[6] = byte(pid >> 8)
	id[7] = byte(pid)
	// 4 bytes for the random value, big endian
    rv := randUint32()
	id[8] = byte(rv >> 24)
	id[9] = byte(rv >> 16)
	id[10] = byte(rv >> 8)
	id[11] = byte(rv)

    // TODO
    // fmt.Println(id.Seconds(), id.Entropy(), len(id)) 

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
// func encode(dst, id []byte) {
// 	encoding.Encode(dst, id[:])
// }


// encode by unrolling the stdlib base32 algorithm + removing all safe checks
func encode(dst, id []byte) {
	_ = dst[19]
	_ = id[11]

	dst[19] = encoding[(id[11]<<4)&0x1F]
	dst[18] = encoding[(id[11]>>1)&0x1F]
	dst[17] = encoding[(id[11]>>6)&0x1F|(id[10]<<2)&0x1F]
	dst[16] = encoding[id[10]>>3]
	dst[15] = encoding[id[9]&0x1F]
	dst[14] = encoding[(id[9]>>5)|(id[8]<<3)&0x1F]
	dst[13] = encoding[(id[8]>>2)&0x1F]
	dst[12] = encoding[id[8]>>7|(id[7]<<1)&0x1F]
	dst[11] = encoding[(id[7]>>4)&0x1F|(id[6]<<4)&0x1F]
	dst[10] = encoding[(id[6]>>1)&0x1F]
	dst[9] = encoding[(id[6]>>6)&0x1F|(id[5]<<2)&0x1F]
	dst[8] = encoding[id[5]>>3]
	dst[7] = encoding[id[4]&0x1F]
	dst[6] = encoding[id[4]>>5|(id[3]<<3)&0x1F]
	dst[5] = encoding[(id[3]>>2)&0x1F]
	dst[4] = encoding[id[3]>>7|(id[2]<<1)&0x1F]
	dst[3] = encoding[(id[2]>>4)&0x1F|(id[1]<<4)&0x1F]
	dst[2] = encoding[(id[1]>>1)&0x1F]
	dst[1] = encoding[(id[1]>>6)&0x1F|(id[0]<<2)&0x1F]
	dst[0] = encoding[id[0]>>3]
}

// Bytes returns by value the byte array representation of ID.
func (id ID) Bytes() []byte {
	return id[:]
}

// Seconds returns the timestamp component of the id in seconds since the Unix
// epoc.
func (id ID) Seconds() int64 {
	// First 4 bytes of ID is 32-bit big-endian seconds from epoch.
	return int64(binary.BigEndian.Uint32(id[0:4]))
}

// Machine returns the 3-byte machine id part of the id.
// It's a runtime error to call this method with an invalid id.
func (id ID) Machine() []byte {
	return id[4:5]
}

// Pid returns the process id part of the id.
// It's a runtime error to call this method with an invalid id.
func (id ID) Pid() uint16 {
	return binary.BigEndian.Uint16(id[5:7])
}

// Time returns the ID's timestamp compoent, with resolution in seconds from
// the Unix epoc.
func (id ID) Time() time.Time {
	return time.Unix(id.Seconds(), 0)
}

// Entropy returns the random component of the ID.
func (id ID) Entropy() uint32 {
    b := id[8:11]
	return uint32(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[0]))
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
	if len(text) != encodedLen {
		return ErrInvalidLength
	}
	for _, c := range text {
        // invalid characters (not in encoding)
		if dec[c] == 0xFF {
			return ErrInvalidID
		}
	}
    // mw's spin
	// count, err := decode(id, text)
	// if (count != rawLen) || (err != nil) {
 //        fmt.Println("here", count, rawLen)
	// 	return ErrInvalidID
	// }
    // rs/xid:
    if !decode(id, text) {
		*id = nilID
		return ErrInvalidID
	}
	return nil
}

// decode a Base32 representation of an ID as a []byte value.
// func decode(id *ID, src []byte) (int, error) {
// 	return encoding.Decode(id[:], src)
// }

// decode by unrolling the stdlib base32 algorithm + customized safe check.
func decode(id *ID, src []byte) bool {
	_ = src[19]
	_ = id[11]

	id[11] = dec[src[17]]<<6 | dec[src[18]]<<1 | dec[src[19]]>>4
	// check the last byte
	if encoding[(id[11]<<4)&0x1F] != src[19] {
		return false
	}
	id[10] = dec[src[16]]<<3 | dec[src[17]]>>2
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

// randUint32 returns a cryptographically secure random uint32
func randUint32() uint32 {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Errorf("rid: cannot generate random number: %v;", err))
	}
    return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func readPlatformMachineID() (string, error) {
	b, err := ioutil.ReadFile("/etc/machine-id")
	if err != nil || len(b) == 0 {
		b, err = ioutil.ReadFile("/sys/class/dmi/id/product_uuid")
	}
    return string(b), err
}

// readMachineId generates machine id and puts it into the machineId global
// variable. If this function fails to get the hostname, it will cause
// a runtime error.
func readMachineID() []byte {
	id := make([]byte, 3)
	hid, err := readPlatformMachineID()
	if err != nil || len(hid) == 0 {
		hid, err = os.Hostname()
	}
	if err == nil && len(hid) != 0 {
		hw := md5.New()
		hw.Write([]byte(hid))
		copy(id, hw.Sum(nil))
	} else {
		// Fallback to rand number if machine id can't be gathered
		if _, randErr := rand.Reader.Read(id); randErr != nil {
			panic(fmt.Errorf("xid: cannot get hostname nor generate a random number: %v; %v", err, randErr))
		}
	}
	return id
}