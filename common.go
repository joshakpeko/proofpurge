// proofpurge is an utility program to help purge financial accounts
// records.
// The goal is to clear matching debit and credit entries pair from
// account records.
package main

import (
    "bytes"
    "encoding/csv"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

var labelRE = regexp.MustCompile(`\d{8}`)
const REindex int = 1

var (
    logBuf bytes.Buffer
    wd, _       = os.Getwd()
    outDir      = filepath.Join(wd, "output")
    logFile     = filepath.Join(outDir, "purge.log")
    debitFile   = filepath.Join(outDir, "debit_out.csv")
    creditFile  = filepath.Join(outDir, "credit_out.csv")
)

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
func (r *Record) isEmpty() bool {
    return len(r.label) == 0
}

type RecordList []*Record

// purge, for each matching records pair (credit and debit), empties
// the corresponding records label.
// It logs purged records into `logBuf`
func (list RecordList) purge() {
    var matches RecordList
    matched := make(map[*Record]bool)

    // matchAllWith adds rec and items in list that matches with rec,
    // to the matches. Note that matches is emptied first.
    matchAllWith := func(rec *Record) {
        matches = nil
        if !matched[rec] && rec.reference != "" {
            matches = append(matches, rec)
            matched[rec] = true
            for _, r := range list {
                // keep all matches, even those in the same list
                if r.isEmpty() || matched[r] {
                    continue
                }
                if r != rec && r.reference == rec.reference {
                    matches = append(matches, r)
                    matched[r] = true
                }
            }
        }
    }
    for _, rec := range list {
        matchAllWith(rec)
        if len(matches) != 2 {
            continue
        }
        // only matches from different lists matter
        if matches[0].debit == matches[1].debit {
            continue
        }
        for _, r := range matches {
            fmt.Fprintf(&logBuf, "%s\n", strings.Join(r.label, ","))
            r.empty()
        }
        fmt.Fprintln(&logBuf)
    }
}

// log logs the last purge() method call details to `logFile`.
// Those details only consist of pairs of records purged.
func (list RecordList) log() error {
    f, err := os.Create(logFile)
    if err != nil {
        return fmt.Errorf("logging: %v", err)
    }
    defer f.Close()
    fmt.Fprintln(f, logBuf.String())
    logBuf.Reset()
    return nil
}

// load reads filename as csv file and return its contents
// in a slice of slices of string.
// If an error occurs during the process, the error-value in non-nil.
func load(filename string) ([][]string, error) {
    f, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("loading %s: %v", filename, err)
    }
    defer f.Close()
    data := csv.NewReader(f)
    return data.ReadAll()
}

// merge takes debit and credit entries and return a RecordList
// containing all entries of the two lists.
func merge(debit, credit [][]string) RecordList {
    var list RecordList
    for i, entry := range debit {
        d := &Record{
            position: i,
            reference: labelRE.FindString(entry[REindex]),
            label: entry,
            debit: true,
        }
        list = append(list, d)
    }
    for i, entry := range credit {
        c := &Record{
            position: i,
            reference: labelRE.FindString(entry[REindex]),
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

    // list holds copies of items in records
    list := make(RecordList, 0, len(records))
    list = append(list, records...)

    for _, rec := range list {
        if !rec.isEmpty() {
            continue
        }
        if rec.debit {
            debit[rec.position] = make([]string, 0)
        } else {
            credit[rec.position] = make([]string, 0)
        }
    }
    // remove empty entries
    cleanDebit := make([][]string, 0, len(debit))
    cleanCredit := make([][]string, 0, len(credit))

    for _, dbt := range(debit) {
        if len(dbt) == 0 {
            continue
        }
        cleanDebit = append(cleanDebit, dbt)
    }

    for _, crd := range(credit) {
        if len(crd) == 0 {
            continue
        }
        cleanCredit = append(cleanCredit, crd)
    }
    return cleanDebit, cleanCredit
}

// dumps writes data to disk as csv file.
// data is considered an account debit entries if debit is true.
// Else, data is considered an account credit entries.
func dump(entries [][]string, debit bool) error {
    var filename string
    if debit {
        filename = debitFile
    } else {
        filename = creditFile
    }
    if err := os.MkdirAll(outDir, 0755); err != nil {
        return fmt.Errorf("dump: %v", err)
    }
    f, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("dump: %v", err)
    }
    defer f.Close()

    data := csv.NewWriter(f)
    if err := data.WriteAll(entries); err != nil {
        return fmt.Errorf("dump: saving %s: %v", filename, err)
    }
    return nil
}
