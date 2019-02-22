// proofpurge is an utility program to help purge financial accounts
// records.
// The goal is to clear matching debit and credit entries pair from
// account records.
package main

import (
    "encoding/csv"
    "os"
    "regexp"
)

var labelRE = regexp.MustCompile(`\d{8}`)
const REindex int = 1

// A Record represents an individual debit or credit record in
// an account journal.
// Each account is consists of two lists: a credit records,
// and a debit records.
// The sequential offset of each record in its corresponding list is
// reported in the `position` field.
// `reference` is computed by scanning the `label` with a regular
// expression. Non-matching scans lead to empty string reference.
// `debit` is the direction of the record: true for debit, false for
// credit.
type Record struct {
    position        int
    reference       string
    label           []string
    debit           bool
}

// empty empties the label field of the r.
func (r *Record) empty() {
    r.label = make([]string, 0)
}

// isEmpty returns true if r's label field is an empty slice, else
// false.
func (r *Record) isEmpty() {
    return len(r.label) == 0
}

type RecordList []*Record

// purge, for each matching records pair (credit and debit), empties
// the corresponding records label.
func (list RecordList) purge() {
    var matches RecordList
    var rec *Record

    visit := func() {
        for _, r := range list {
            if r.isEmpty() || r.debit == rec.debit {
                continue
            }
            matches = append(matches, r)
        }
    }
    for i := range list {
        rec = list[i]
        matches = nil
        matches = append(matches, rec)
        visit()
        if len(matches) != 2 {
            continue
        }
        for _, r := range match {
            r.empty()
        }
    }
}

// load reads filename as csv file and return its contents
// in a slice of slices of string.
// If an error occurs during the process, the error-value in non-nil.
func load(filename string) ([][]string, error) {
    f, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    data := csv.NewReader(f)
    return data.ReadAll()
}

// merge takes debit and credit entries and return a RecordList
// containing all the entries of the two lists.
func merge(debit, credit [][]string) RecordList {
    var list RecordList
    for i, entry := range debit {
        d := &Record{
            position: i,
            reference: labelRE.FindString(entry[REindex])
            label: entry,
            debit: true,
        }
        list = append(list, d)
    }
    for i, entry := range credit {
        c := &Record{
            position: i,
            reference: labelRE.FindString(entry[REindex])
            label: entry,
        }
        list = append(list, c)
    }
    return list
}

// for each individual entry in debit and credit, mirror looks
// for the corresponding record in records, and then reflects changes
// made to latter in the former.
// It then returns the mirrored debit and credit entries.
func mirror(records RecordList, debit, credit [][]string) ([][]string, [][]string) {
    for _, rec := range records {
        if !rec.isEmpty() {
            continue
        }
        if rec.debit {
            debit[rec.position] = make([]string, 0)
        } else {
            credit[rec.position] = make([]string, 0)
        }
    }
    return debit, credit
}

// dumps writes data to disk as csv file.
// data is considered an account debit entries if debit is true.
// Else, data is considered an account credit entries.
func dump(entries [][]string, debit bool) error {
    var filename string
    if debit {
        filename = "debit_out.csv"
    } else {
        filename = "credit_out.csv"
    }
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()
    data := csv.NewWriter(f)
    return data.WriteAll(entries)
}
