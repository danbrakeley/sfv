package sfv

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func Test_CreateFromFile(t *testing.T) {
	cases := []struct {
		Name string
	}{
		{"winsfv3211a"},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			inName := filepath.Join("testdata", tc.Name+".sfv")
			goldenName := filepath.Join("testdata", tc.Name+".golden")

			sf, err := CreateFromFile(inName)
			if err != nil {
				t.Fatalf("error reading %s: %v", inName, err)
			}

			var actual string
			for i, entry := range sf.Files {
				actual += fmt.Sprintf("%2d %s %s\n", i, hex.EncodeToString(entry.CRC32), entry.Filename)
			}

			if *update {
				os.WriteFile(goldenName, []byte(actual), 0o644)
			}

			expected, _ := os.ReadFile(goldenName)
			if !bytes.Equal([]byte(actual), expected) {
				t.Fatalf(
					"golden file %s does not match test output\nexpected:\n%s\nactual:\n%s\n",
					goldenName, expected, actual,
				)
			}
		})
	}
}

func Test_Verify(t *testing.T) {
	cases := []struct {
		Name string
	}{
		{"verify-ok"},
		{"verify-mismatch"},
		{"verify-missing-file"},
		{"verify-missing-sfv-file"},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			inName := filepath.Join("testdata", tc.Name+".sfv")
			goldenName := filepath.Join("testdata", tc.Name+".golden")
			var actual string

			sf, err := CreateFromFile(inName)
			if err != nil {
				actual += fmt.Sprintf("%v\n", err)
			}

			// test with a nil progress func
			results := sf.Verify()
			for i, entry := range results.Files {
				actual += fmt.Sprintf("%d %s %s %s %v\n", i, entry.ExpectedCRC32, entry.ActualCRC32, entry.Filename, entry.Err)
			}

			// test with a non-nil progress func
			fnProgress := func(file string, read, total int64) {
				if total > 0 && read > total {
					t.Errorf("fnProgress called with invalid byte count %d/%d", read, total)
				}
				if len(file) == 0 {
					t.Errorf("fnProgress called with empty file name")
				}
				actual += fmt.Sprintf("%d/%d %s\n", read, total, file)
			}

			results = sf.Verify(fnProgress)
			for i, entry := range results.Files {
				actual += fmt.Sprintf("%d %s %s %s %v\n", i, entry.ExpectedCRC32, entry.ActualCRC32, entry.Filename, entry.Err)
			}

			if *update {
				os.WriteFile(goldenName, []byte(actual), 0o644)
			}

			expected, _ := os.ReadFile(goldenName)
			if !bytes.Equal([]byte(actual), expected) {
				t.Fatalf(
					"golden file %s does not match test output\nexpected:\n%s\nactual:\n%s\n",
					goldenName, expected, actual,
				)
			}
		})
	}
}
