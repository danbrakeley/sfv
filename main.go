package main

import (
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/danbrakeley/frog"
)

func main() {
	var flagAll = flag.Bool("all", false, "check all sfv files in the current directory")
	var flagJSON = flag.Bool("json", false, "log output as json")
	flag.Parse()

	var log frog.Logger
	if *flagJSON {
		log = frog.New(frog.JSON)
	} else {
		log = frog.New(frog.Auto)
	}
	defer log.Close()

	var sfvFiles []string
	if *flagAll {
		files, err := ioutil.ReadDir("./")
		if err != nil {
			log.Fatal("unable to read current directory", frog.Err(err))
		}
		for _, f := range files {
			name := f.Name()
			if !f.IsDir() && strings.HasSuffix(name, ".sfv") {
				sfvFiles = append(sfvFiles, name)
			}
		}
		if len(sfvFiles) == 0 {
			log.Warning("no sfv files found")
			return
		}
	} else if len(flag.Args()) == 1 {
		arg := flag.Arg(0)
		if info, err := os.Stat(arg); os.IsNotExist(err) {
			log.Fatal("unable to stat file", frog.String("path", arg), frog.Err(err))
		} else if info.IsDir() {
			log.Fatal("path is a directory, not a file", frog.String("path", arg))
		}
		sfvFiles = append(sfvFiles, arg)
	}

	if len(sfvFiles) == 0 {
		log.Fatal(fmt.Sprintf("usage: %s [-json] [-all | <file.sfv>]", filepath.Base(os.Args[0])))
	}

	hasBadCRC := false
	for _, sfvFile := range sfvFiles {
		sfvs, err := ReadSFVFile(sfvFile)
		if err != nil {
			log.Fatal("error reading sfv file", frog.String("sfv", sfvFile), frog.Err(err))
		}

		for _, sfv := range sfvs {
			hash, err := GenerateCRC32ForFile(sfv.Filename)
			if err != nil {
				log.Fatal("error reading file", frog.String("file", sfv.Filename), frog.Err(err), frog.String("sfv", sfvFile))
			}

			if hash != sfv.CRC32 {
				hasBadCRC = true
				log.Error(
					"CRC32 mismatch",
					frog.String("file", sfv.Filename),
					frog.String("expected", fmt.Sprintf("%08x", sfv.CRC32)),
					frog.String("actual", fmt.Sprintf("%08x", hash)),
					frog.String("sfv", sfvFile),
				)
			} else {
				log.Info(
					"OK",
					frog.String("file", sfv.Filename),
					frog.String("crc32", fmt.Sprintf("%08x", sfv.CRC32)),
					frog.String("sfv", sfvFile),
				)
			}
		}
	}

	if hasBadCRC {
		log.Close()
		os.Exit(1)
	}
}

func GenerateCRC32ForFile(filename string) (uint32, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	hash := crc32.NewIEEE()
	if _, err := io.Copy(hash, file); err != nil {
		return 0, err
	}

	return hash.Sum32(), nil
}
