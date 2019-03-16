// proofpurge is an utility program to help purge financial account
// records.
// The goal is to clear matching debit-side and credit-side entries
// that are conterpart of one another, for a given account.
package main

import (
    "bytes"
    "encoding/csv"
    "errors"
    "flag"
    "fmt"
    "io"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"
)

var (
    specREMin  = regexp.MustCompile(
        fmt.Sprintf(`\d{%d,}?`, *refSize),
    )
    // strings that also match specREMax are not valid ref strings.
    specREMax  = regexp.MustCompile(
        fmt.Sprintf(`\d{%d,}`, *refSize),
    )
    refSize     = flag.Int(
        "refsize", 12, `length of the reference string`,
    )
    specCol      = flag.Int(
        "speccolumn", 1, `column number of the specification text
        field (text describing the transaction)`,
    )
    amountCol   = flag.Int(
        "amountcolumn", 2, `column number of the amount field`,
    )
    dateCol     = flag.Int(
        "datecolumn", 0, `column number of the date field`,
    )
)

// A Record represents an individual debit or credit entry in
// an account journal.
// Each account consists of two lists: credit-side entries and
// debit-side entries.
type Record struct {
    // offset is the sequential number of the record in its
    // corresponding list (debit or credit)
    offset          int
    // label is the entire entry line.
    label           []string
    // debit specifies the direction of the record:
    // true for debit-side entry, false for credit-side entry.
    debit           bool
    // date is the record's register date.
    // It serves optimization purpose: avoid repeated computations.
    date            time.Time
}

// Clear empties the label field of the r.
func (r *Record) Clear() {
    r.label = make([]string, 0)
}

// IsEmpty returns true if r's label field is an empty slice, else
// false.
func (r *Record) IsEmpty() bool {
    return len(r.label) == 0
}

// Ref computes and returns the reference of r.
// This reference represents the subset of r's specification text
// that matches the regular expression specREMin but not specREMax.
// An empty string means that r has no reference.
func (r *Record) Ref() string {
    if r.IsEmpty() || len(r.label) <= *specCol {
        return ""
    }
    s := r.label[*specCol]
    min := specREMin.FindString(s)
    if max := specREMax.FindString(s); max != min {
        return ""
    }
    return min
}

// Amount returns a float64 representing the amount of the record.
// P.S. amount in term of money sum representing the value of the
// transaction.
// Amount returns also an error value in case there's been some issues
// when parsing the amount as float64.
func (r *Record) Amount() (float64, error) {
    if r.IsEmpty() || len(r.label) <= *amountCol {
        return 0, errors.New("*Record.Amount: no amount field found")
    }
    return strconv.ParseFloat(r.label[*amountCol], 64)
}

// Date locates, parses and returns the registration date of r.
// If no valid date was found, the result is Jan 1 year 0.
func (r *Record) Date() time.Time {
    var t time.Time
    if len(r.label) > *dateCol {
        t, _ = time.Parse("02/01/06", r.label[*dateCol])
    }
    return t
}

// RecordListMap is a mapping of a common string to a list of
// records that share that string.
type RecordListMap struct {
    // refMap maps references text to records.
    refMap      map[string][]*Record
    // specMap maps specification texts to records.
    specMap     map[string][]*Record
    // number of debit-side entries.
    nDebit      int
    // number of credit-side entries.
    nCredit     int
    buf         bytes.Buffer
}

// Add adds a new record to m.
func (m *RecordListMap) Add(r *Record) {
    ref := r.Ref()
    spec := r.label[*specCol]
    if m.refMap == nil {
        m.refMap = make(map[string][]*Record)
    }
    if m.specMap == nil {
        m.specMap = make(map[string][]*Record)
    }
    m.refMap[ref] = append(m.refMap[ref], r)
    m.specMap[spec] = append(m.specMap[spec], r)

    if r.debit {
        m.nDebit++
    } else {
        m.nCredit++
    }
}

// AddAll adds all items to m.
func (m *RecordListMap) AddAll(items ...*Record) {
    for _, r := range items {
        m.Add(r)
    }
}

// Purge clears matching records.
// 'clear' here means calling the corresponding method of *Record on
// records that are counterparts of one another.
// Note that 2 records are considered counterparts, if they share
// either the same reference string, or the same specification text.
// Purge returns the number of records cleared.
func (m *RecordListMap) Purge() int {
    var count int
    purge := func(mapping map[string][]*Record) {
        for k, v := range mapping {
            if k == "" {
                continue
            }
            tm := TrueMatch(v...)
            for _, r := range tm {
                fmt.Fprintf(
                    &m.buf, "%s\n", strings.Join(r.label, ","),
                )
                r.Clear()
                count++
            }
            if len(tm) > 0 {
                fmt.Fprintln(&m.buf, "")    // add gap in logs.
            }
        }
    }
    purge(m.specMap)
    purge(m.refMap)
    return count
}

// Pack returns two separate groups of data representing:
// for the 1st, the labels of records from m that are debit entries,
// for the 2nd, the labels of records that are credit entries.
func (m *RecordListMap) Pack() ([][]string, [][]string) {

    debit := make([][]string, m.nDebit)
    credit := make([][]string, m.nCredit)

    for _, v := range m.refMap {
        for _, r := range v {
            if r.debit {
                debit[r.offset] = r.label
            } else {
                credit[r.offset] = r.label
            }
        }
    }
    debitCompact := make([][]string, 0, m.nDebit)
    creditCompact := make([][]string, 0, m.nDebit)

    // remove empty lines.
    for _, d := range debit {
        if len(d) > 0 {
            debitCompact = append(debitCompact, d)
        }
    }
    for _, c := range credit {
        if len(c) > 0 {
            creditCompact = append(creditCompact, c)
        }
    }
    return debitCompact, creditCompact
}

// Log writes the details of the last `purge` method calls on list to
// w. These details only consist of pairs of records purged since
// the last time this method was called.
// Note that each time log is called, list.buf is flushed.
// Log returns any write error encountered.
func (m *RecordListMap) Log(w io.Writer) error {
    _, err := fmt.Fprintln(w, m.buf.String())
    m.buf.Reset()
    return err
}

// RecQueue is an implementation of a Queue of *Record.
// It's actually a named slice of *Record with additional methods.
type RecQueue struct {
    recs []*Record
}

func (q *RecQueue) EnQueue(rec *Record) {
    q.recs = append(q.recs, rec)
}

func (q *RecQueue) DeQueue() (*Record, bool) {
    if q.IsEmpty() {
        return nil, false
    }
    r := q.recs[0]
    q.recs = q.recs[1:]
    return r, true
}

func (q *RecQueue) Size() int {
    return len(q.recs)
}

func (q *RecQueue) IsEmpty() bool {
    return len(q.recs) == 0
}

func (q *RecQueue) Clear() {
    q.recs = make([]*Record, 0)
}

// TrueMatch returns a slice of recs from items that are real matches.
// Two records are considered true matches when:
// 1. they share the same reference or spec (taken for granted here)
// 2. one is a debit entry and the other is a credit entry
// 3. the have the same money amount.
// TrueMatch consider all items to share the same reference or
// specification text, so caller should make sure it is the case
// prior to calling TrueMatch.
func TrueMatch(items ...*Record) []*Record {

    result := make([]*Record, 0, len(items))

    recs := make([]*Record, len(items))
    copy(recs, items)
    SortByDate(recs)

    // m maps amounts to records
    m := make(map[float64][]*Record)
    for _, rec := range recs {
        amount, err := rec.Amount()
        if err != nil {
            continue
        }
        m[amount] = append(m[amount], rec)
    }
    /*
    *q := new(RecQueue)
    *for _, v := range m {
        *q.Clear()
        *for _, r := range v {
            *if r.debit {
                *q.EnQueue(r)
            *} else {
                *if q.IsEmpty() {
                    *continue
                *}
                *rr, _ := q.DeQueue()
                *result = append(result, rr)
                *result = append(result, r)
            *}
        *}
    *}
    */
    dq := new(RecQueue)     // debit recs queue
    cq := new(RecQueue)     // credit recs queue
    for _, v := range m {
        cq.Clear()
        dq.Clear()
        for _, r := range v {
            if r.debit {
                dq.EnQueue(r)
            } else {
                cq.EnQueue(r)
            }
        }
        for !dq.IsEmpty() && !cq.IsEmpty() {
            drec, _ := dq.DeQueue()
            result = append(result, drec)
            crec, _ := cq.DeQueue()
            result = append(result, crec)
        }

    }
    return result
}

// NewRecord builds an returns a valid *Record instance
//from given inputs.
func NewRecord(s []string, offset int, debit bool) *Record {
    r := &Record{
        offset: offset,
        label: s,
        debit: debit,
    }
    r.date = r.Date()
    return r
}

// SortByDate sorts a slice of *Record by date.
func SortByDate(recs []*Record) {
    sort.SliceStable(recs, func(i, j int) bool {
        return recs[i].date.Before(recs[j].date)
    })
}
// Load reads the content of r as a csv formatted data.
// It returns the result as a slice of slices of string.
// If an error occurs during the process, the error-value in non-nil.
func Load(r io.Reader) ([][]string, error) {
    data := csv.NewReader(r)
    return data.ReadAll()
}

// Dumps writes data as a csv file to the writer w.
// data is considered an account's debit entries if debit is true.
// Else, data is considered an account's credit entries.
func Dump(w io.Writer, entries [][]string) error {
    data := csv.NewWriter(w)
    return data.WriteAll(entries)
}
