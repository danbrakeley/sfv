# sfv

## Overview

A package for parsing sfv files, as well as verifying the listed checksums against actual local files.

The root is the sfv package itself, and also included in this repo are cli and gui apps (as separate modules, see `cmd/*`) for verifying sfv checksums and viewing the results.

## Example Usage

To open and parse an sfv file encoded in utf8:

```go
  sf, err := sfv.CreateFromFile("testfile.sfv")
  if err != nil {
    panic(err)
  }

  results := sf.Verify()

  // print the results...
  for _, entry := range results.Files {
    if len(entry.Err) == 0 {
      fmt.Println(entry.Filename + " checks out")
    } else {
      fmt.Println(entry.Filename + " does not match the sfv checksum of " + entry.ExpectedCRC32)
    }
  }

  // ...and/or show the results as a json object
  b, err := json.MarshalIndent(results, "", "  ")
  if err != nil {
    panic(err)
  }
  fmt.Println(string(b))
```

Note that when calling Verify, you can optionally include 1 or more funcs that are called after each file is checksummed, to allow for incremental progress to be displayed.

```go
  results := sf.Verify(func(file string, bytesRead, bytesTotal int64) {
    fmt.Printf("Processing... %3d%%\n", 100*bytesRead/bytesTotal)
  })
  fmt.Println("Done!")
```
