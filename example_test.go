// Package sid_test provides examples for godoc/pkg.dev
package sid_test

import (
	"fmt"
	"time"

	"github.com/solutionroute/sid"
)

func ExampleNew() {
	id := sid.New()
	fmt.Printf(`ID:
    String()       %s   
    Milliseconds() %d  
    Count()        %d // random for this one-off run 
    Time()         %v
    Bytes():       %3v  
`, id.String(), id.Milliseconds(), id.Count(), id.Time(), id.Bytes())
}

func ExampleNewWithTime() {
	id := sid.NewWithTime(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC))
	fmt.Printf(`ID:
    String()       %s
    Milliseconds() %d
    Count()        %d // random for this one-off run 
    Time()         %v
    Bytes():       %3v
`, id.String(), id.Milliseconds(), id.Count(), id.Time().UTC(), id.Bytes())
}

func ExampleFromString() {
	id, err := sid.FromString("af1z631jaa0y4")
	if err != nil {
		panic(err)
	}
	fmt.Printf(`ID:
    String()       %s
    Milliseconds() %d
    Count()        %d
    Time()         %v
    Bytes():       %3v
`, id.String(), id.Milliseconds(), id.Count(), id.Time().UTC(), id.Bytes())
}
