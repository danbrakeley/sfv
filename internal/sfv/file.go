package sfv

import (
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

type File struct {
	Filename string      `json:"filename"`
	Files    []FileEntry `json:"files"`
}

type FileEntry struct {
	Filename string `json:"name"`
	CRC32    uint32 `json:"crc"`
}

type VerifyResults struct {
	RootPath string
	Files    []ResultsEntry
}

type ResultsEntry struct {
	Filename      string
	ExpectedCRC32 uint32
	ActualCRC32   uint32
	ActualSize    int64
	Err           error
}

func CreateFromFile(filename string) (File, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return File{}, err
	}

	if !utf8.Valid(b) {
		return File{}, fmt.Errorf("sfv file is not valid utf8")
	}

	sf := File{
		Filename: filename,
	}

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
			return File{}, fmt.Errorf("line %d has non-utf8 filename", i)
		}
		rawCRC := line[len(line)-8:]
		if !utf8.Valid([]byte(rawCRC)) {
			return File{}, fmt.Errorf("line %d has non-utf8 crc32", i)
		}
		parsedCRC, err := strconv.ParseUint(rawCRC, 16, 32)
		if err != nil {
			return File{}, fmt.Errorf("error parsing crc32: %v", err)
		}

		sf.Files = append(sf.Files, FileEntry{Filename: rawFilename, CRC32: uint32(parsedCRC)})
	}

	return sf, nil
}

func (sf File) Verify(fnProgress func(curFile string, bytesRead, bytesTotal int64)) VerifyResults {
	// filenames in the sfv will be relative to the sfv file itself
	rootPath := filepath.Dir(sf.Filename)

	results := VerifyResults{
		RootPath: rootPath,
		Files:    make([]ResultsEntry, len(sf.Files)),
	}

	var bytesTotal int64
	for i, entry := range sf.Files {
		results.Files[i].Filename = entry.Filename
		results.Files[i].ExpectedCRC32 = entry.CRC32

		if fi, err := os.Stat(filepath.Join(rootPath, entry.Filename)); err != nil {
			results.Files[i].Err = err
		} else if fi.IsDir() {
			results.Files[i].Err = fmt.Errorf("name refers to a directory, not a file")
		} else {
			s := fi.Size()
			results.Files[i].ActualSize = s
			bytesTotal += s
		}
	}

	var bytesRead int64
	for i := range results.Files {
		re := &results.Files[i]

		if re.Err != nil {
			continue
		}

		if fnProgress != nil {
			fnProgress(re.Filename, bytesRead, bytesTotal)
		}

		hash, err := GenerateCRC32ForFile(filepath.Join(rootPath, re.Filename))
		bytesRead += re.ActualSize
		if err != nil {
			re.Err = err
			continue
		}

		re.ActualCRC32 = hash
		if re.ActualCRC32 != re.ExpectedCRC32 {
			re.Err = fmt.Errorf("checksum mismatch")
			continue
		}
	}

	return results
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
