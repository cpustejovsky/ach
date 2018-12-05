// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

// Package main is an example for creating an Automated Clearing House (ACH) file with Moov's HTTP service.
// To run this example first start the ach service locally:
//   $ go run ./cmd/server // from this project's root directory
//
// Then, in a second terminal you can run this example:
//   $ go run ./examples/http // from project root
package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/moov-io/ach"
)

var (
	// achAddress refers to the local host and port for the ACH service running locally
	achAddress = "http://localhost:8080"
)

func main() {
	// Example transfer to write an ACH PPD file to send/credit a external institutions account
	// Important: All financial institutions are different and will require registration and exact field values.

	// Set originator bank ODFI and destination Operator for the financial institution
	// this is the funding/receiving source of the transfer
	fh := ach.NewFileHeader()
	fh.ImmediateDestination = "231380104" // Routing Number of the ACH Operator or receiving point to which the file is being sent
	fh.ImmediateOrigin = "121042882"      // Routing Number of the ACH Operator or sending point that is sending the file
	fh.FileCreationDate = time.Now()      // Today's Date
	fh.ImmediateDestinationName = "Federal Reserve Bank"
	fh.ImmediateOriginName = "My Bank Name"

	// BatchHeader identifies the originating entity and the type of transactions contained in the batch
	bh := ach.NewBatchHeader()
	bh.ServiceClassCode = 220          // ACH credit pushes money out, 225 debits/pulls money in.
	bh.CompanyName = "Name on Account" // The name of the company/person that has relationship with receiver
	bh.CompanyIdentification = fh.ImmediateOrigin
	bh.StandardEntryClassCode = "PPD"         // Consumer destination vs Company CCD
	bh.CompanyEntryDescription = "REG.SALARY" // will be on receiving accounts statement
	bh.EffectiveEntryDate = time.Now().AddDate(0, 0, 1)
	bh.ODFIIdentification = "121042882" // Originating Routing Number

	// Identifies the receivers account information
	// can be multiple entry's per batch
	entry := ach.NewEntryDetail()
	// Identifies the entry as a debit and credit entry AND to what type of account (Savings, DDA, Loan, GL)
	entry.TransactionCode = 22          // Code 22: Credit (deposit) to checking account
	entry.SetRDFI("231380104")          // Receivers bank transit routing number
	entry.DFIAccountNumber = "12345678" // Receivers bank account number
	entry.Amount = 100000000            // Amount of transaction with no decimal. One dollar and eleven cents = 111
	entry.SetTraceNumber(bh.ODFIIdentification, 1)
	entry.IndividualName = "Receiver Account Name" // Identifies the receiver of the transaction

	// build the batch
	batch := ach.NewBatchPPD(bh)
	batch.AddEntry(entry)
	if err := batch.Create(); err != nil {
		log.Fatalf("Unexpected error building batch: %s\n", err)
	}

	// build the file
	file := ach.NewFile()
	file.SetHeader(fh)
	file.AddBatch(batch)
	if err := file.Create(); err != nil {
		log.Fatalf("Unexpected error building file: %s\n", err)
	}

	// Encode our ACH File as JSON for the upload...
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&file); err != nil {
		log.Fatal(err)
	}

	// Make our HTTP request to the ACH service
	req, err := http.NewRequest("POST", achAddress+"/files/create", &buf)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Printf("File created!")
	} else {
		bs, _ := ioutil.ReadAll(resp.Body)
		log.Fatalf("error creating file: %v", string(bs))
	}
}