package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"
)

var (
    wd, _       = os.Getwd()
    outDir      = filepath.Join(wd, "purge")
    logFile     = filepath.Join(outDir, "purge.log")
    debitOfn    = filepath.Join(outDir, "debit_out.csv")
    creditOfn   = filepath.Join(outDir, "credit_out.csv")
)

var (
    cfn         = flag.String("cf", "", `credit csv filename`)
    dfn         = flag.String("df", "", `debit csv filename`)
)

const templ = `
`

func main() {
    start := time.Now()
    flag.Parse()

    if len(*cfn) == 0 || len(*dfn) == 0 {
        os.Exit(1)
    }

    log.SetPrefix("proofpurge: ")
    fmt.Println("processing ...\n")

    // loads
    debit, err := loadDebit()
    if err != nil {
        log.Fatal(err)
    }
    credit, err := loadCredit()
    if err != nil {
        log.Fatal(err)
    }

    var rmap RecordListMap
    addto(&rmap, true, debit...)
    addto(&rmap, false, credit...)

    npurged := rmap.Purge()

    if err = os.MkdirAll(outDir, 0755); err != nil {
        log.Fatal(fmt.Sprintf("creating destination folder: %v", err))
    }

    // logs
    if err := logPurged(&rmap); err != nil {
        log.Fatal(err)
    }

    debit, credit = rmap.Pack()

    // saves
    if err := saveDebit(debit); err != nil {
        log.Fatal(err)
    }
    if err := saveCredit(credit); err != nil {
        log.Fatal(err)
    }

    elapsed := time.Since(start)

    // prints
    //fmt.Println("Done!")
    fmt.Printf(
        "%d records successfully purged in %v.\n", npurged, elapsed,
    )
    fmt.Printf("results saved to %s/ :\n", filepath.Base(outDir))
    fmt.Printf(
        "files written on %s and %s\n",
        filepath.Base(debitOfn), filepath.Base(creditOfn),
    )
    fmt.Printf(
        "transcript written on %s.\n", filepath.Base(logFile),
    )
}

func loadDebit() ([][]string, error) {
    f, err := os.Open(*dfn)
    if err != nil {
        return nil, fmt.Errorf("loading: %v", err)
    }
    defer f.Close()
    return Load(f)
}

func loadCredit() ([][]string, error) {
    f, err := os.Open(*cfn)
    if err != nil {
        return nil, fmt.Errorf("loading: %v", err)
    }
    defer f.Close()
    return Load(f)
}

func addto(m *RecordListMap, debit bool, items ...[]string) {
    for i, item := range items {
        m.Add(NewRecord(item, i, debit))
    }
}

func logPurged(m *RecordListMap) error {
    f, err := os.Create(logFile)
    if err != nil {
        return fmt.Errorf("logging: %v", err)
    }
    defer f.Close()
    return m.Log(f)
}

func saveDebit(data [][]string) error {
    f, err := os.Create(debitOfn)
    if err != nil {
        return fmt.Errorf("saving: %v", err)
    }
    defer f.Close()
    return Dump(f, data)
}

func saveCredit(data [][]string) error {
    f, err := os.Create(creditOfn)
    if err != nil {
        return fmt.Errorf("saving: %v", err)
    }
    defer f.Close()
    return Dump(f, data)
}
