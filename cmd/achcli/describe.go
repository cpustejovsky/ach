// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/moov-io/ach"
	"github.com/moov-io/ach/cmd/achcli/describe"
)

func dumpFiles(paths []string, validateOpts *ach.ValidateOpts) error {
	var files []*ach.File
	for i := range paths {
		f, err := readACHFile(paths[i], validateOpts)
		if err != nil {
			fmt.Printf("WARN: problem reading %s:\n %v\n\n", paths[i], err)
		}
		files = append(files, f)
	}

	if *flagMerge {
		merged, err := ach.MergeFiles(files)
		if err != nil {
			fmt.Printf("ERROR: merging files: %v\n", err)
		}
		fmt.Printf("Describing %d file(s) merged into %d file(s)\n", len(paths), len(merged))
		files = merged
	}

	if *flagFlatten {
		for i := range files {
			fmt.Printf("attempting flattening %d\n", i)
			file, err := files[i].FlattenBatches()
			if err != nil {
				fmt.Printf("ERROR: problem flattening file: %v\n", err)
			}
			files[i] = file
		}
	}

	for i := range files {
		if i > 0 && len(files) > 1 {
			fmt.Println("") // extra newline between multiple ACH files
		}
		if !*flagMerge {
			fmt.Printf("Describing ACH file '%s'\n\n", paths[i])
		}
		if files[i] != nil {
			describe.File(os.Stdout, files[i], &describe.Opts{
				MaskAccountNumbers: *flagMask || *flagMaskAccounts,
				MaskCorrectedData:  *flagMask || *flagMaskCorrectedData,
				MaskNames:          *flagMask || *flagMaskNames,
				PrettyAmounts:      *flagPretty || *flagPrettyAmounts,
			})
		} else {
			fmt.Printf("nil ACH file in position %d\n", i)
		}
	}

	return nil
}

func readACHFile(path string, validateOpts *ach.ValidateOpts) (*ach.File, error) {
	fd, readErr := os.Open(path)
	if readErr != nil {
		return nil, fmt.Errorf("problem opening %s: %v", path, readErr)
	}
	defer fd.Close()

	r := ach.NewReader(fd)
	r.SetValidation(validateOpts)
	f, err := r.Read()
	if err != nil {
		return nil, err
	}
	return &f, err
}
