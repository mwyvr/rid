package sid

import (
	"database/sql/driver"
	enc "encoding"
	"reflect"
	"sync"
	"testing"
	"time"
)

var (
	now = time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

	// for testing concurrency safety
	wg sync.WaitGroup
)

type idTest struct {
	name         string
	valid        bool
	id           ID
	rawbytes     []byte
	milliseconds uint64
	counter      uint16
	b32          string
}

var testIDS = []idTest{
	{
		"min value 1970-01-01 00:00:00 +0000 UTC", // epoch time
		true,
		ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		[]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		0,
		0,
		"aaaaaaaaaaaaa",
	},
	{
		"max value in the year 10889 see you then",
		true,
		ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		281474976710655,
		65535,
		"9999999999998",
	},
	{
		"fail on FromString / FromBytes - value mismatch",
		false,
		ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xaa},
		281474976710655,
		65535,
		"abbadabba9998",
	},
	{
		"fail on FromString, FromBytes len mismatch",
		false,
		ID{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		[]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		281474976710655,
		65535,
		"abbadabba",
	},
	{
		"Jan 1 2020, year of the pandemic",
		true,
		ID{1, 111, 94, 102, 232, 0, 77, 75},
		[]byte{1, 111, 94, 102, 232, 0, 77, 75},
		1577836800000,
		19787,
		"af1z631jabgy0",
	},
}

func TestNew(t *testing.T) {
	counter = 0 // package var
	for i := 1; i <= 1000; i++ {
		_ = New()
	}
	if counter != 1000 {
		t.Errorf("counter at %d, should be 1000", counter)
	}
}

func TestCounterRollover(t *testing.T) {
	counter = 65534 // package var
	id := New()     // 65535, then set to zero
	id = New()      // counter now at 1
	if (counter != 1) || (id.Count() != 1) {
		t.Errorf("counter at %d, should be 0", counter)
	}
}

func TestNew_Unique(t *testing.T) {
	var d = &dupes{count: make(map[string]int)}
	// generate 4,999,990 IDs concurrently
	// load it up... no failures observed at *much* higher loads
	numConcurrent := 10 // go routines X
	numIter := 500000   // id creation
	for i := 1; i <= numConcurrent; i++ {
		wg.Add(1)
		go func() {
			for i := 1; i < numIter; i++ {
				id := New()
				d.add(id.String())
			}
			wg.Done()
		}()
	}
	wg.Wait()
	d.report(t)
}

func TestNewWithTime(t *testing.T) {
	type args struct {
		tm time.Time
	}
	tests := []struct {
		name string
		args args
		want ID
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewWithTime(tt.args.tm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewWithTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

type sidTest struct {
	name         string
	valid        bool
	id           ID
	milliseconds uint64
	counter      uint16
	b32          string
}

func TestID_Components(t *testing.T) {
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.String(); (got != tt.b32) && (tt.valid != false) {
				t.Errorf("ID.String() = %v, want %v", got, tt.b32)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Milliseconds(); (got != tt.milliseconds) && (tt.valid != false) {
				t.Errorf("ID.Milliseconds() = %v, want %v", got, tt.milliseconds)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Count(); (got != tt.counter) && (tt.valid != false) {
				t.Errorf("ID.Count() = %v, want %v", got, tt.counter)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Bytes(); !reflect.DeepEqual(got, tt.id[:]) {
				t.Errorf("ID.Bytes() = %#v, want %#v", got, tt.id)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Milliseconds(); got != tt.milliseconds {
				t.Errorf("ID.Milliseconds() = %v, want %v", got, tt.milliseconds)
			}
		})
	}
}
func TestID_Time(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want time.Time
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.Time(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ID.Time() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromString(t *testing.T) {
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromString(tt.b32)
			if tt.valid && (err != nil) {
				t.Errorf("FromString() error = %v, is valid %v", err, tt.valid)
				return
			}
			_ = err
			if tt.valid && !reflect.DeepEqual(got, tt.id) {
				t.Errorf("FromString() = %v, want %v", got, tt.id)
			}
		})
	}
}

func TestFromBytes(t *testing.T) {
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromBytes(tt.rawbytes)
			if tt.valid && (err != nil) {
				t.Errorf("FromBytes() error = %v, wantErr %v", err, tt.valid)
				return
			}
			if tt.valid && !reflect.DeepEqual(got, tt.id) {
				t.Errorf("FromBytes() = %v, want %v", got, tt.id)
			}
		})
	}
}

func Test_encode(t *testing.T) {
	for _, tt := range testIDS {
		t.Run(tt.name, func(t *testing.T) {
			dest := make([]byte, encodedLen)
			encode(dest, tt.id[:])
			if len(dest) != encodedLen {
				t.Errorf("encode: wrong length got %d want %d", len(dest), encodedLen)
				return
			}
			if tt.valid && string(dest) != tt.b32 {
				t.Errorf("encode: wrong output got %s want %s", dest, tt.b32)
			}
		})
	}
}

func Test_decode(t *testing.T) {
	type args struct {
		buf []byte
		src []byte
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decode(tt.args.buf, tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("decode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_randInt(t *testing.T) {
	tests := []struct {
		name string
		want uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := randInt(); got != tt.want {
				t.Errorf("randInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestID_UnmarshalText(t *testing.T) {
	type args struct {
		text []byte
	}
	tests := []struct {
		name    string
		id      *ID
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		// ensure ID implements the interface
		var _ enc.TextUnmarshaler = &ID{}
		t.Run(tt.name, func(t *testing.T) {
			// Ensure that ID fulfills the Interface
			var _ enc.TextUnmarshaler = &ID{}

			if err := tt.id.UnmarshalText(tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("ID.UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestID_MarshalText(t *testing.T) {
	tests := []struct {
		name    string
		id      ID
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ensure ID implements the interface
			var _ enc.TextMarshaler = &ID{}
			got, err := tt.id.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("ID.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ID.MarshalText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestID_Value(t *testing.T) {
	tests := []struct {
		name    string
		id      ID
		want    driver.Value
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.id.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("ID.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ID.Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestID_Scan(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		id      *ID
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.id.Scan(tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("ID.Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// mutext protected map/counter for checking for uniqueness
type dupes struct {
	count map[string]int
	mu    sync.Mutex
}

func (d *dupes) add(str string) {
	d.mu.Lock()
	d.count[str]++
	d.mu.Unlock()
}

func (d *dupes) report(t *testing.T) {
	d.mu.Lock()
	defer d.mu.Unlock()
	var total = 0
	for v, num := range d.count {
		if num > 1 {
			total++
			id, err := FromString(v)
			if err != nil {
				t.Errorf("id.FromString: %s, %s", v, err)
			}
			if id == nilID {
				t.Errorf("id.FromString produced nilID: %s, %s", v, err)
			}
		}
	}
	if total != 0 {
		// there should be zero
		t.Errorf("Duplicate Base36 values (keys) found. Total dupes: %d | Total keys: %d\n",
			total, len(d.count))
	}
}
