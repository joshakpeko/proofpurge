// proofpurge is an utility program to help purge financial accounts
// records.
// The goal is to clear matching debit and credit entries pair from
// account records.
package main

import (
    "bytes"
    "encoding/csv"
    "fmt"
    "io"
    //"os"
    //"path/filepath"
    "regexp"
    "strings"
    //"time"
)

var labelRE = regexp.MustCompile(`\d{8}`)
const REindex int = 1

// A Record represents an individual debit or credit record in
// an account journal.
// Each account consists of two lists: a credit records,
// and a debit records.
// The sequential offset of each record in its corresponding list is
// reported in the `offset` field.
// `debit` is the direction of the record: true for debit, false for
// credit.
type Record struct {
    offset          int
    label           []string
    debit           bool
}

// Empty empties the label field of the r.
func (r *Record) Empty() {
    r.label = make([]string, 0)
}

// IsEmpty returns true if r's label field is an empty slice, else
// false.
func (r *Record) IsEmpty() bool {
    return len(r.label) == 0
}

// Ref computes and returns the reference of r.
// This reference represents the subset of r.label that matches the
// regular expression labelRE.
// Note that the result may be an empty string, in case no
// match has been found.
func (r *Record) Ref() string {
    if r.IsEmpty() || len(r.label) <= REindex {
        return ""
    }
    return labelRE.FindString(r.label[REindex])
}

// Record returs true if r matches s, meaning r and s have the same
// references.
// It returns false otherwise.
func  (r *Record) Match(s *Record) bool {
    if ref := r.Ref(); ref != "" && ref == s.Ref() {
        return true
    }
    return false
}

// RecordList keeps a list of *Records.
// It also has a memory buffer useful for recording purged Records.
type RecordList struct {
    buf         bytes.Buffer
    Records     []*Record
}

// Purge, for each matching records pair (credit and debit), empties
// the corresponding records label.
// It logs purged records into the memory buffer in RecordList,
// and returns the number of records pair purged.
func (list *RecordList) Purge() int {
    var count int
    var matches []*Record
    matched := make(map[*Record]bool)

    // matchAllWith adds `rec` plus each item in list that matches
    // with it to `matches`. Note that matches is emptied at each run.
    matchAllWith := func(rec *Record) {
        matches = nil
        if !matched[rec] && rec.Ref() != "" {
            matches = append(matches, rec)
            matched[rec] = true
            for _, r := range list.Records {
                // keep all matches, even those in the same list
                if !matched[r] && rec.Match(r) {
                    matches = append(matches, r)
                    matched[r] = true
                }
            }
        }
    }
    for _, rec := range list.Records {
        if matchAllWith(rec); len(matches) != 2 {
            continue
        }
        // only matches from different lists matter
        if matches[0].debit == matches[1].debit {
            continue
        }
        for _, r := range matches {
            fmt.Fprintf(&list.buf, "%s\n", strings.Join(r.label, ","))
            r.Empty()
        }
        fmt.Fprintln(&list.buf)
        count++
    }
    return count
}

// Log writes the details of the last `purge` method calls of list to
// w. These details only consist of pairs of records purged since
// the last time this method was called.
// Note that each time log is called, the memory buffer
// in the `buf` field of list is emptied.
// Log returns any write error encountered.
func (list *RecordList) Log(w io.Writer) error {
    _, err := fmt.Fprintln(w, list.buf.String())
    list.buf.Reset()
    return err
}

// Load reads the content of r as a csv formatted data.
// It returns the result as a slice of slices of string.
// If an error occurs during the process, the error-value in non-nil.
func Load(r io.Reader) ([][]string, error) {
    data := csv.NewReader(r)
    return data.ReadAll()
}

// Merge takes debit and credit entries and return a RecordList
// containing all entries of the two lists converted into 
// *Record entities.
func Merge(debit, credit [][]string) *RecordList {
    var list RecordList
    for i, entry := range debit {
        d := &Record{
            offset: i,
            label: entry,
            debit: true,
        }
        list.Records = append(list.Records, d)
    }
    for i, entry := range credit {
        c := &Record{
            offset: i,
            label: entry,
        }
        list.Records = append(list.Records, c)
    }
    return &list
}

// for each individual entry in debit and credit, Mirror looks
// for the corresponding record in records, and then reflects changes
// made to latter in the former.
// It then returns the mirrored debit and credit entries.
// Note that original debit and credit provided by the caller are not
// modified.
func Mirror(records *RecordList, debit, credit [][]string) ([][]string, [][]string) {

    // list holds copies of items in records
    list := make([]*Record, 0, len(records.Records))
    list = append(list, records.Records...)

    for _, rec := range list {
        if !rec.IsEmpty() {
            continue
        }
        if rec.debit {
            debit[rec.offset] = make([]string, 0)
        } else {
            credit[rec.offset] = make([]string, 0)
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

// Dumps writes data as a csv file to the writer w.
// data is considered an account's debit entries if debit is true.
// Else, data is considered an account's credit entries.
func Dump(w io.Writer, entries [][]string) error {
    data := csv.NewWriter(w)
    return data.WriteAll(entries)
}
