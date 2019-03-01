package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
)

var (
    wd, _       = os.Getwd()
    outDir      = filepath.Join(wd, "output")
    logFile     = filepath.Join(outDir, "purge.log")
    debitFile   = filepath.Join(outDir, "debit_out.csv")
    creditFile  = filepath.Join(outDir, "credit_out.csv")
)

var crd = flag.String("c", "", `debit csv filename`)
var dbt = flag.String("d", "", `credit csv filename`)

func main() {
    flag.Parse()
    if len(*crd) == 0 || len(*dbt) == 0 {
        os.Exit(1)
    }
    log.SetPrefix("proofpurge: ")

    var npurged int

    // load csv files
    crdfile, err := os.Open(*crd)
    if err != nil {
        log.Fatal(fmt.Sprintf("loading file: %v", err))
    }
    defer crdfile.Close()

    dbtfile, err := os.Open(*dbt)
    if err != nil {
        log.Fatal(fmt.Sprintf("loading file: %v", err))
    }
    defer crdfile.Close()

    var credit, debit [][]string

    credit, err = Load(crdfile)
    if err != nil {
        log.Fatal(err)
    }
    debit, err = Load(dbtfile)
    if err != nil {
        log.Fatal(err)
    }

    // merge all account records
    records := Merge(debit, credit)
    if err != nil {
        log.Fatal(err)
    }

    npurged += records.Purge()

    debit, credit = Mirror(records, debit, credit)

    // save purged records to csv files
    if err = os.MkdirAll(outDir, 0755); err != nil {
        log.Fatal(fmt.Sprintf("creating destination folder: %v", err))
    }
    df, err := os.Create(debitFile)
    if err != nil {
        log.Fatal(fmt.Sprintf("writing %s: %v", debitFile, err))
    }
    defer df.Close()

    cf, err := os.Create(creditFile)
    if err != nil {
        log.Fatal(fmt.Sprintf("writing %s: %v", creditFile, err))
    }
    defer cf.Close()

    if err = Dump(df, debit); err != nil {
        log.Fatal(fmt.Sprintf("writing %s: %v", debitFile, err))
    }
    if err = Dump(cf, credit); err != nil {
        log.Fatal(fmt.Sprintf("writing %s: %v", creditFile, err))
    }

    // create and fill log file
    lf, err := os.Create(logFile)
    if err != nil {
        log.Fatal(fmt.Sprintf("writing %s: %v", logFile, err))
    }
    defer lf.Close()
    if err = records.Log(lf); err != nil {
        log.Fatal(fmt.Sprintf("writing %s: %v", logFile, err))
    }

    /*...ouput results to user...*/

    fmt.Println("Done!")
    fmt.Printf("%d records pair(s) successfully purged.\n", npurged)
    fmt.Printf("Results saved to %s/ :\n", filepath.Base(outDir))

    fmt.Printf("files written on %s and %s\n",
    filepath.Base(debitFile), filepath.Base(creditFile))

    fmt.Printf("transcript written on %s.\n", filepath.Base(logFile))
}
