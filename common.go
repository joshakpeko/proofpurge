// proofpurge package provide utilities to purge financial accounts
// entries.
// The goal is to clear matching debit and credit entry pairs from
// account summary.
package proofpurge

import (
    "encoding/csv"
    "regexp"
)

// An accountEntry represents an operation entry line in an account
// journal.
// Each account is first represented two lists: a summary of debit
// entries, and a summary of credit entries.
// Thus the position in accountEntry is the corresponding sequential
// order number of the entry in the corresponding list.
// reference is computed by scanning the label with a regular
// expression. Non-matching scans lead to empty string reference.
// debit is the direction of the entry: true for debit, false for
// credit.
type accountEntry struct {
    position        int
    reference       string
    label           []string
    debit           bool
}

// empty empties the label field of the e.
// This actually means e should be considered as deleted from entries.
func (e *accountEntry) empty() {
    e.label = make([]string, 0)
}

type entryList []*accountEntry

// purge, for each matching entry pairs (credit and debit), empties
// the corresponding entries label.
func (list *entryList) purge() {
}
