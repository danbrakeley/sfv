package sfv

import (
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type File struct {
	Filename string
	Files    []FileEntry
}

type FileEntry struct {
	Filename string
	CRC32    []byte
}

// VerifyResults holds the results of calling Verify on a given sfv.File object.
// This object is built to be easy to parse when marshalled to JSON.
type VerifyResults struct {
	SFVFile string         `json:"sfv"`
	Files   []ResultsEntry `json:"files"`
}

type ResultsEntry struct {
	Filename      string `json:"filename"`
	ExpectedCRC32 string `json:"expected_crc"`
	ActualCRC32   string `json:"actual_crc,omitempty"`
	ActualSize    int64  `json:"size,omitempty"`
	Err           string `json:"error,omitempty"`
}

// CreateFromFile parses the given file into an sfv.File object.
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
		b, err := hex.DecodeString(rawCRC)
		if err != nil {
			return File{}, fmt.Errorf("error parsing crc32: %v", err)
		}

		sf.Files = append(sf.Files, FileEntry{Filename: rawFilename, CRC32: b})
	}

	return sf, nil
}

// Verify will open each file, compute the checksum, then add the results to the returned VerifyResults object.
func (sf File) Verify(fnProgress func(curFile string, bytesRead, bytesTotal int64)) VerifyResults {
	// filenames in the sfv will be relative to the sfv file itself
	rootPath := filepath.Dir(sf.Filename)

	results := VerifyResults{
		SFVFile: filepath.ToSlash(filepath.Clean(sf.Filename)),
		Files:   make([]ResultsEntry, len(sf.Files)),
	}

	var bytesTotal int64
	for i, entry := range sf.Files {
		results.Files[i].Filename = entry.Filename
		results.Files[i].ExpectedCRC32 = hex.EncodeToString(entry.CRC32)

		if fi, err := os.Stat(filepath.Join(rootPath, entry.Filename)); err != nil {
			results.Files[i].Err = err.Error()
		} else if fi.IsDir() {
			results.Files[i].Err = "name refers to a directory, not a file"
		} else {
			s := fi.Size()
			results.Files[i].ActualSize = s
			bytesTotal += s
		}
	}

	var bytesRead int64
	for i := range results.Files {
		re := &results.Files[i]

		if len(re.Err) != 0 {
			continue
		}

		if fnProgress != nil {
			fnProgress(re.Filename, bytesRead, bytesTotal)
		}

		hash, err := GenerateCRC32ForFile(filepath.Join(rootPath, re.Filename))
		bytesRead += re.ActualSize
		if err != nil {
			re.Err = err.Error()
			continue
		}

		re.ActualCRC32 = fmt.Sprintf("%08x", hash)
		if re.ActualCRC32 != re.ExpectedCRC32 {
			re.Err = "checksum mismatch"
			continue
		}
	}

	return results
}

// GenerateCRC32ForFile is a helper to open a file, compute its checksum, then close the file.
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
