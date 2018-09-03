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
