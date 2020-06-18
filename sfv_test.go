package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func Test_ReadSFVFile(t *testing.T) {
	cases := []struct {
		Name string
	}{
		{"winsfv3211a"},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			sfvFile := filepath.Join("testdata", tc.Name+".sfv")
			goldenFile := filepath.Join("testdata", tc.Name+".golden")

			sfvs, err := ReadSFVFile(sfvFile)
			if err != nil {
				t.Fatalf("error reading %s: %v", sfvFile, err)
			}

			var actual bytes.Buffer
			for _, sfv := range sfvs {
				actual.WriteString(sfv.Filename)
				actual.WriteString("\n")
				actual.WriteString(fmt.Sprintf("%08x\n", sfv.CRC32))
			}

			if *update {
				ioutil.WriteFile(goldenFile, actual.Bytes(), 0644)
			}

			expected, _ := ioutil.ReadFile(goldenFile)
			if !bytes.Equal(actual.Bytes(), expected) {
				t.Fatalf(
					"golden file %s does not match test output\nexpected:\n%s\nactual:\n%s\n",
					goldenFile, expected, actual.String(),
				)
			}
		})
	}
}
