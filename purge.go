package main

import (
    "flag"
    "fmt"
    "log"
    "os"
)


var c = flag.String("c", "", `debit csv filename`)
var d = flag.String("d", "", `credit csv filename`)

func main() {
    flag.Parse()
    var credit, debit [][]string

    // load csv files
    credit, err := load(*c)
    if err != nil {
        log.Fatal(err)
    }
    debit, err := load(*d)
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
    }
    if err = dump(credit, true); err != nil {
        log.Fatal(err)
    }
}
