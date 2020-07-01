package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/danbrakeley/frog"
	"github.com/danbrakeley/sfv"
)

func main() {
	var flagJSON = flag.Bool("json", false, "output results as json")
	flag.Parse()

	var log frog.Logger
	if *flagJSON {
		log = frog.New(frog.Basic, frog.HideTimestamps, frog.FieldIndent20)
	} else {
		log = frog.New(frog.Auto, frog.HideTimestamps, frog.FieldIndent20)
	}
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

	if *flagJSON {
		b, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			log.Fatal("error generating json", frog.Err(err))
		}
		fmt.Print(string(b))
		return
	}

	hasErrors := false
	for _, entry := range results.Files {
		if len(entry.Err) == 0 {
			log.Info("OK", frog.String("file", entry.Filename), frog.String("crc", entry.ActualCRC32))
		} else {
			log.Error("mismatch!",
				frog.String("file", entry.Filename),
				frog.String("expected_crc", entry.ExpectedCRC32),
				frog.String("actual_crc", entry.ActualCRC32),
				frog.String("error", entry.Err),
			)
			hasErrors = true
		}
	}

	if hasErrors {
		log.Close()
		os.Exit(-1)
	}
}
