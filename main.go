package main

import (
    "encoding/csv"
    "flag"
    "fmt"
    "log"
    "os"
    "regexp"
)


var c = flag.String("c", "", `debit csv filename`)
var d = flag.String("d", "", `credit csv filename`)

func main() {
    flag.Parse()
    var credits, debits [][]string

    // load csv files
    credits, err := load(*c)
    if err != nil {
        log.Fatal(err)
    }
    debits, err := load(*d)
    if err != nil {
        log.Fatal(err)
    }

    // merge all account entries
    entries, err := merge(credits, debits)
    if err != nil {
        log.Fatal(err)
    }

    entries.purge()

    if err = mirror(entries, debits, credits); err != nil {
        log.Fatal(err)
    }

    // save purged entries to csv files
    if err = dump(debits, true); err != nil {
        log.Fatal(err)
    }
    }
    if err = dump(credits, true); err != nil {
        log.Fatal(err)
    }
}
