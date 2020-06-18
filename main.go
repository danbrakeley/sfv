package main

import (
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
)

func main() {
	flag.Parse()

	var sfvFiles []string
	for _, arg := range flag.Args() {
		if info, err := os.Stat(arg); os.IsNotExist(err) {
			fmt.Printf("error opening '%s': %v\n", arg, err)
			os.Exit(-1)
		} else if info.IsDir() {
			fmt.Printf("error: '%s' is a directory, not a file\n", arg)
			os.Exit(-1)
		}

		sfvFiles = append(sfvFiles, arg)
	}

	if len(sfvFiles) == 0 {
		fmt.Printf("no sfv files found\n")
		fmt.Printf("usage: %s sfv-file [ sfv-file [...]]\n", filepath.Base(os.Args[0]))
		os.Exit(-1)
	}

	for _, sfvFile := range sfvFiles {
		sfvs, err := ReadSFVFile(sfvFile)
		if err != nil {
			fmt.Printf("error reading '%s': %v\n", sfvFile, err)
			os.Exit(-1)
		}

		for _, sfv := range sfvs {
			hash, err := GenerateCRC32ForFile(sfv.Filename)
			if err != nil {
				fmt.Printf("error processing '%s': %v\n", sfv.Filename, err)
				os.Exit(-1)
			}

			if hash != sfv.CRC32 {
				fmt.Printf("%s: expected %08x, got %08x\n", sfv.Filename, sfv.CRC32, hash)
			} else {
				fmt.Printf("%s: OK!\n", sfv.Filename)
			}
		}
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
