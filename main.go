// Proofpurge... The functional way...
package main

import (
    "fmt"
    "regexp"
    "sort"
    "strconv"
    "sync"
    "time"
)

// ?: -> non-capturing group
const refRegString = `(?:([^\d]+|^))(?P<%s>\d{%d})(?:([^\d]+|$))`

var (
    reFlag = "ref"
    defaultRefSize = 8
)

func main() {
    /*
    formattedRE := fmt.Sprintf(refRegString, reFlag, defaultRefSize)
    refRE := regexp.MustCompile(formattedRE)
    */

    // testing
    /*
    raw1 := [][]string{
        {
            "16/12/19",
            "alibaba12345678& les 40 voleurs",
            "250000",
        },
        {
            "15/09/19",
            "asterix20 and Ob46364546lix 309",
            "35000",
        },
    }
    raw2 := [][]string{
        {
            "16/08/19",
            "12345678 for more...",
            "250000",
        },
        {
            "15/09/19",
            "asterix20 and Ob12345678 309",
            "250000",
        },
        {
            "17/11/19",
            "asterix20 and Ob46364546lix 309",
            "35000",
        },
    }
    fmt.Println(raw1)
    fmt.Println(raw2)
    fmt.Println()

    group1 := toEntryGroup( toEntryList(raw1, refRE, reFlag) )
    group2 := toEntryGroup( toEntryList(raw2, refRE, reFlag) )

    mGroup1, mGroup2 := markMatch(group1, group2)

    clean1 := prune(raw1, markedIdList(mGroup1))
    clean2 := prune(raw2, markedIdList(mGroup2))

    trash := matchedEntriesList(mGroup1, mGroup2)
    for _, l := range trash {
        for _, entry := range l {
            fmt.Printf("%#v\n", entry.line)
        }
        fmt.Println()
    }

    fmt.Println(clean1)
    fmt.Println(clean2)
    */

    /*
    entries1 := toEntryList(raw1, refRE, reFlag)
    entries2 := toEntryList(raw2, refRE, reFlag)

    group1 := toEntryGroup(entries1)
    group2 := toEntryGroup(entries2)

    out1, out2 := markMatch(group1, group2)

    ids1 := markedIdList(out1)
    ids2 := markedIdList(out2)

    clean1 := prune(raw1, ids1)
    clean2 := prune(raw2, ids2)
    */

    /*
    */


    /*
    fmt.Println(ids1)
    fmt.Println(ids2)
    for k, v := range out1 {
        fmt.Printf("%s: ", k)
        for _, entry := range v {
            fmt.Printf("%#v\n", entry)
        }
    }
    fmt.Println()
    for k, v := range out2 {
        fmt.Printf("%s: ", k)
        for _, entry := range v {
            fmt.Printf("%#v\n", entry)
        }
    }
    for _, entry := range entries1 {
        fmt.Printf("%#v\n", entry)
    }
    for _, entry := range entries2 {
        fmt.Printf("%#v\n", entry)
    }
    fmt.Println()
    out1, out2 := matchWith(entries1, entries2)
    for _, entry := range out1 {
        fmt.Printf("%#v\n", entry)
    }
    for _, entry := range out2 {
        fmt.Printf("%#v\n", entry)
    }
    */
}

type Id int

// A record in an account registry.
type accountEntry struct {
    idx    Id
    ref    string
    descr  string // description
    value  float64
    date   time.Time
    line   []string
    marked bool // entry to be pruned or not
}

/* ==>> FROM [][]string TO []*accountEntry <<== */

func toEntryList(
    raw [][]string, re *regexp.Regexp, reFlag string,
) []*accountEntry {

    n := len(raw)
    list := make([]*accountEntry, 0, n)
    ch := make(chan *accountEntry)

    getEntry := func(r []string, i int, out chan<- *accountEntry) {
        entry, err := toAccountEntry(r, i, re, reFlag)
        if err != nil {
            return
        }
        ch <- &entry
    }
    var wg sync.WaitGroup
    for i, r := range raw {
        wg.Add(1)
        go func(r []string, idx int) {
            defer wg.Done(); getEntry(r, idx, ch)
        }(r, i)
    }
    go func() {
        wg.Wait()
        close(ch)
    }()
    for entry := range ch {
        list = append(list, entry)
    }
    return list
}

func toAccountEntry(
    raw []string, idx int,
    re *regexp.Regexp, reFlag string,
) (accountEntry, error) {

    nfield, dateIdx, descrIdx, valueIdx := 3, 0, 1, 2
    n := len(raw)
    if n < nfield {
        return accountEntry{}, fmt.Errorf(
            "not enough field: %v, has %d, expecting %d",
            raw, n, nfield,
        )
    }

    var descr string
    if d, err := getDescr(raw, descrIdx); err != nil {
        return accountEntry{}, err
    } else {
        descr = d
    }

    var value float64
    if v, err := getValue(raw, valueIdx); err != nil {
        return accountEntry{}, err
    } else {
        value = v
    }

    var date time.Time
    if d, err := getDate(raw, dateIdx); err != nil {
        return accountEntry{}, err
    } else {
        date = d
    }

    return accountEntry{
        idx:    Id(idx),
        ref:    getRef(descr, re, reFlag),
        descr:  descr,
        value:  value,
        date:   date,
        line:   raw,
        marked: false,
    }, nil
}

func getDescr(raw []string, idx int) (string, error) {
    if len(raw) < idx {
        return "", noFieldError("description", raw)
    }
    return raw[idx], nil
}

func getRef(s string, re *regexp.Regexp, reFlag string) string {
    names := re.SubexpNames()
    matches := re.FindStringSubmatch(s)
    if matches == nil {
        return ""
    }
    for i := range matches {
        if names[i] == reFlag {
            return matches[i]
        }
    }
    return ""
    //return re.FindString(s)
}

func getValue(raw []string, idx int) (float64, error) {
    if len(raw) < idx {
        return 0, noFieldError("value", raw)
    }
    value, err := strconv.ParseFloat(raw[idx], 64)
    if err != nil {
        return 0, fieldParsingError("value", err)
    }
    return value, nil
}

func getDate(raw []string, idx int) (time.Time, error) {
    if len(raw) < idx {
        return time.Date(
            0, 0, 0, 0, 0, 0, 0, time.Local,
        ), noFieldError("date", raw)
    }
    t, err := time.Parse("02/01/06", raw[idx])
    if err != nil {
        return t, fieldParsingError("date", err)
    }
    return t, nil
}

func noFieldError(fieldname string, raw []string) error {
    return fmt.Errorf("no %s field found: %v", fieldname, raw)
}

func fieldParsingError(fieldname string, err error) error {
    return fmt.Errorf("parsing %s field: %v", fieldname, err)
}

/* ==>> FROM []*accountEntry TO map[string][]*accountEntry <<== */

// entryGroup map ref of descr (fields) to a slice of accountEntry.
type entryGroup map[string][]*accountEntry

func toEntryGroup(entries []*accountEntry) entryGroup {
    mapping := make(entryGroup)
    for _, entry := range entries {
        mapping = addToGroup(mapping, entry)
    }
    return mapping
}

// Add *accountry to a group in accountGroup based on either its ref
// or its descr field.
func addToGroup(group entryGroup, entry *accountEntry) entryGroup {
    var key string
    if entry.ref != "" {
        key = entry.ref
    } else {
        key = entry.descr
    }
    entries := group[key]
    entries = append(entries, entry)
    group[key] = entries
    return group
}

/* ==>> FROM entryGroup TO `MARKED` entryGroup <<== */

// markMatch identifies entries in both grp1 and grp2 that match
// together (same ref or desc && same value).
// It returns new groups where matched entries have their marked field
// set to true.
func markMatch(grp1, grp2 entryGroup) (out1, out2 entryGroup) {
    for k, v := range grp1 {
        list1 := v
        var list2 []*accountEntry

        if l, ok := grp2[k]; !ok {
            continue
        } else {
            list2 = l
        }
        grp1[k], grp2[k] = matchWith(list1, list2)
    }
    return grp1, grp2
}

// matchWith matches entries in both lists and returns lists in
// with matched entries have their marked field set to true.
func matchWith(l1, l2 []*accountEntry) (out1, out2 []*accountEntry) {
    var subgrp1, subgrp2 map[string][]*accountEntry
    subgrp1 = groupByValue(l1)
    subgrp2 = groupByValue(l2)
    done := make(chan struct{})

    // sort within both subgroups concurrently
    go func() {
        defer func() { done <- struct{}{} }()
        for k, v := range subgrp1 {
            subgrp1[k] = sortByDate(v)
        }
    }()
    go func() {
        defer func() { done <- struct{}{} }()
        for k, v := range subgrp2 {
            subgrp2[k] = sortByDate(v)
        }
    }()
    for i := 0; i < 2; i++ {
        <-done
    }

    for k, v := range subgrp1 {
        list1 := v
        var list2 []*accountEntry

        if l, ok := subgrp2[k]; !ok {
            continue
        } else {
            list2 = l
        }

        subgrp1[k], subgrp2[k] = markMore(list1, list2)
    }
    return flatten(subgrp1), flatten(subgrp2)
}

func groupByValue(l []*accountEntry) map[string][]*accountEntry {
    m := make(map[string][]*accountEntry)
    for _, entry := range l {
        key := strconv.FormatFloat(entry.value, 'f', -1, 64)
        m[key] = append(m[key], entry)
    }
    return m
}

func sortByDate(l []*accountEntry) []*accountEntry {
    sort.SliceStable(
        l,
        func(i, j int) bool {
            date1 := l[i].date
            date2 := l[j].date
            return date1.Before(date2)
        },
    )
    return l
}

// markMore marks the maximum possible same number of entries in both
// lists and returns the result.
func markMore(l1, l2 []*accountEntry) (out1, ou2 []*accountEntry) {
    least := min(len(l1), len(l2))
    done := make(chan struct{})

    var m1, m2 []*accountEntry

    // mark items within both list concurrently
    go func() {
        defer func() { done <- struct{}{} }()
        m1 = markAll(l1[:least])
    }()
    go func() {
        defer func() { done <- struct{}{} }()
        m2 = markAll(l2[:least])
    }()
    for i := 0; i < 2; i++ {
        <-done
    }

    // add non-marked items from old lists to the new ones.
    if len(m1) < len(l1) {
        m1 = append(m1, l1[least:]...)
    }
    if len(m2) < len(l2) {
        m2 = append(m2, l2[least:]...)
    }

    return m1, m2
}

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

// markAll sets the marked field of entries within lst to true, and
// returns the result.
func markAll(lst []*accountEntry) []*accountEntry {
    n := len(lst)
    done := make(chan struct{}, n)

    for _, entry := range lst {
        go func() {
            defer func() { done <- struct{}{} }()
            entry.marked = true
        }()
    }

    for i := 0; i < n; i++ {
        <-done
    }
    return lst
}

func flatten(m map[string][]*accountEntry) []*accountEntry {
    var list []*accountEntry
    for _, v := range m {
        list = append(list, v...)
    }
    return list
}

/* ==>> FROM entryGroup TO []Id <<== */

// markedIdList extracts and returns Ids from items that are marked in
// group.
func markedIdList(group entryGroup) []Id {
    return toIdList(markedOnly(flatten(group)))
}

// toIdList returns a sorted list of Ids of entries from lst.
func toIdList(lst []*accountEntry) []Id {
    idList := make([]Id, 0, len(lst))
    for _, entry := range lst {
        idList = append(idList, entry.idx)
    }
    return idList
}

// markedOnly returns a new list composed exclusively of entry that
// have their marked field set to true.
func markedOnly(lst []*accountEntry) []*accountEntry {
    newList := make([]*accountEntry, 0, len(lst))
    for _, entry := range lst {
        if entry.marked {
            newList = append(newList, entry)
        }
    }
    return newList
}

/* ==>>
* FROM `MARKED` entryGroup TO MATCH PAIRS [][]*accountEntry
* <<== */

// matchedEntriesList bundles together entries that have been matched
// with each others and returns a list of these bundles.
func matchedEntriesList(grp1, grp2 entryGroup) [][]*accountEntry {
    var bundles [][]*accountEntry
    ch := make(chan []*accountEntry)
    var wg sync.WaitGroup

    for k, v := range grp1 {
        lst1 := v
        var lst2 []*accountEntry
        if l, ok := grp2[k]; !ok {
            continue
        } else {
            lst2 = l
        }

        wg.Add(1)
        go func(lst1, lst2 []*accountEntry) {
            defer wg.Done()
            ch <- bundle(lst1, lst2)
        }(lst1, lst2)
    }
    go func() {
        wg.Wait()
        close(ch)
    }()
    for entries := range ch {
        bundles = append(bundles, entries)
    }
    return bundles
}

// bundle returns a list of marked entries with the same value in both
// lists.
func bundle(lst1, lst2 []*accountEntry) []*accountEntry {
    var list []*accountEntry
    subgrp1 := groupByValue(lst1)
    subgrp2 := groupByValue(lst2)

    for k, v := range subgrp1 {
        entries1 := v
        var entries2 []*accountEntry
        if l, ok := subgrp2[k]; !ok {
            continue
        } else {
            entries2 = l
        }
        list = append(list, markedOnly(entries1)...)
        list = append(list, markedOnly(entries2)...)
    }
    return list
}

/* ==>> FROM [][]string TO `PRUNED` [][]string <<== */

// prune returns a new list composed of elements in items that don't
// have their index listed in idList.
func prune(items [][]string, idList []Id) [][]string {
    sort.SliceStable(
        idList,
        func(i, j int) bool {
            return int(idList[i]) < int(idList[j])
        },
    )
    var out [][]string
    for i, item := range items {
        switch {
        case len(idList) == 0:
            out = append(out, item)
        case i == int(idList[0]):
            idList = idList[1:] // pop head
        default:
            out = append(out, item)
        }
    }
    return out
}
