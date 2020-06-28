# sfv

## Overview

Verify files against checksums from the given sfv file.

## Usage

`sfv [-json] <sfv-filename>`

argument | description
--- | ---
`-json` | output summary is valid json object; checksum mismatches do not cause a non-zero exit code
`<sfv-filename>` | sfv file to process
