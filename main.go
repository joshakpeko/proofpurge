package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
)


var crd = flag.String("c", "", `debit csv filename`)
var dbt = flag.String("d", "", `credit csv filename`)

func main() {
    flag.Parse()
    if len(*crd) == 0 || len(*dbt) == 0 {
        os.Exit(1)
    }
    var credit, debit [][]string
    log.SetPrefix("proofpurge: ")

    // load csv files
    credit, err := load(*crd)
    if err != nil {
        log.Fatal(err)
    }
    debit, err = load(*dbt)
    if err != nil {
        log.Fatal(err)
    }

    // merge all account records
    records := merge(debit, credit)
    if err != nil {
        log.Fatal(err)
    }

    records.purge()

    debit, credit = mirror(records, debit, credit)

    // save purged records to csv files
    if err = dump(debit, true); err != nil {
        log.Fatal(err)
    }
    if err = dump(credit, false); err != nil {
        log.Fatal(err)
    }
    if err = records.log(); err != nil {
        fmt.Fprintf(os.Stderr, "%v\n", err)
    }

    /*...ouput results to user...*/

    fmt.Println("Done!")
    fmt.Printf("Results saved to %s/ :\n", filepath.Base(outDir))

    fmt.Printf("\tfiles written on %s and %s.\n",
    filepath.Base(debitFile), filepath.Base(creditFile))

    fmt.Printf("\ttranscript written on %s.\n", filepath.Base(logFile))
}
