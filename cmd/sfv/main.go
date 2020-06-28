package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/danbrakeley/frog"
	"github.com/danbrakeley/sfv/internal/sfv"
)

func main() {
	// var flagJSON = flag.Bool("json", false, "output results as json")
	flag.Parse()

	log := frog.New(frog.Auto)
	defer log.Close()

	files := flag.Args()
	if len(files) != 1 {
		log.Fatal(fmt.Sprintf("usage: %s [-json] <file.sfv>", filepath.Base(os.Args[0])))
	}

	line := frog.AddFixedLine(log)
	line.Transient(fmt.Sprintf("Parsing %s...", files[0]))

	sf, err := sfv.CreateFromFile(files[0])
	if err != nil {
		log.Fatal("error parsing sfv file", frog.Err(err), frog.String("file", files[0]))
	}

	fnProgress := func(filename string, read, total int64) {
		line.Transient(fmt.Sprintf("Checking %s %3d%%", filename, 100*read/total))
	}

	results := sf.Verify(fnProgress)
	frog.RemoveFixedLine(line)

	hasErrors := false
	for _, entry := range results.Files {
		if entry.Err == nil {
			log.Info("OK", frog.String("file", entry.Filename), frog.String("crc", fmt.Sprintf("%08X", entry.ActualCRC32)))
		} else {
			log.Error("mismatch!",
				frog.String("file", entry.Filename),
				frog.String("expected_crc", fmt.Sprintf("%08X", entry.ExpectedCRC32)),
				frog.String("actual_crc", fmt.Sprintf("%08X", entry.ActualCRC32)),
				frog.Err(entry.Err),
			)
			hasErrors = true
		}
	}

	if hasErrors {
		log.Close()
		os.Exit(-1)
	}
}
