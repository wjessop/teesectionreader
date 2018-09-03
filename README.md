# TeeSectionReader

TeeSectionReader is like an amalgamation of the StdLib SectionReader, and the TeeReader. Things to note about it are:

1. It takes a writer as an argument to the create function
2. It writes all bytes read to the writer provided

The original goal of this was to be able to checksum the data read using the Read method by passing in an object that is a writer and passing all data read to this writer.

This is complicated by the fact that the some readers (AWS SDK for instance) will read the source data twice, once to checksum a request, and once to transfer the data. There was no easy way of getting round this so I've had to implement a sort of "ratchet" writer, keeping track of the offset of data written so that re-reads aren't then sent to the writer, messing up the checksums.

## Usage

```go
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	tsr "github.com/wjessop/teesectionreader"
)

func main() {
	input := bytes.NewReader([]byte("Digest this!"))

	output, err := os.Create("/tmp/outputfile")
	if err != nil {
		panic(err)
	}

	defer output.Close()

	// Loop over subsequent sections of the input and output them
	for i := 0; i < 3; i++ {
		// When we read from input the contents read will also be written to output
		sectionReader := tsr.NewTeeSectionReader(input, output, int64(i)*4, 4)
		b, err := ioutil.ReadAll(sectionReader)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(b))
	}

	// You will find the contents that were read were also written to the output file
	//
	// $ cat /tmp/outputfile
	// Digest this!
}
```


