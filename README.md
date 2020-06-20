# sfv

## Overview

Load an sfv file, then for each file listed in the sfv, find the corresponding local file, calculate that file's crc32, compare with the crc32 in the sfv, and output the results.

## Usage

`sfv [-json] [-all | <sfv-filename>]`

argument | description
--- | ---
`-json` | format output as json, one valid json object per line of output
`-all` | find and process all files ending in `.sfv` in the current directory
`<sfv-filename>` | name of sfv file to process

exit code | description
--- | ---
`0` | all CRC32s match
`1` | any single file does not match
`-1` | unrecoverable error during run
