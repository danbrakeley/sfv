package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"unicode/utf8"
)

type SFVEntry struct {
	Filename string
	CRC32    uint32
}

func ReadSFVFile(filename string) ([]SFVEntry, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return []SFVEntry{}, err
	}

	if !utf8.Valid(b) {
		return []SFVEntry{}, fmt.Errorf("sfv file is not valid utf8")
	}

	var out []SFVEntry

	for i, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if len(line) < 10 {
			continue
		}

		if line[0] == ';' {
			continue
		}

		rawFilename := line[:len(line)-9]
		if !utf8.Valid([]byte(rawFilename)) {
			return []SFVEntry{}, fmt.Errorf("line %d has non-utf8 filename", i)
		}
		rawCRC := line[len(line)-8:]
		if !utf8.Valid([]byte(rawCRC)) {
			return []SFVEntry{}, fmt.Errorf("line %d has non-utf8 crc32", i)
		}
		parsedCRC, err := strconv.ParseUint(rawCRC, 16, 32)
		if err != nil {
			return []SFVEntry{}, fmt.Errorf("error parsing crc32: %v", err)
		}

		out = append(out, SFVEntry{Filename: rawFilename, CRC32: uint32(parsedCRC)})
	}

	return out, nil
}
