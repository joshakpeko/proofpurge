# proofpurge

Book keeping help tool.

## Description

proofpurge is a financial account purger.

Book keepers record each financial transaction in corresponding
accounts. And in double-side accounting, an operation usually has its
counterpart, meaning another operation that goes the other way. And so
goes for records on each account involved in those operations.

For financial account involved in such operations, we need a way to
keep track of `debit` records that don't yet have `credit` records
counterpart or vice versa. That is, we must `purge` each account of
records that already have counterpart, leaving the summary alone
with pending records. The summary is usually called proof; hence
the name: proofpurge.

To purge actually means systematically clearing records from the
proof. And two records are considered counterparts of one another if:

  - one is a debit-side entry and the other a credit-side entry
  - they both have the same reference or the same specification text
  - they have the same amount.

This long and tedious purge process is usually done with spreadsheets.
But it's still a pain and time-consuming job to harness the power and
simplicity of spreadsheets to get this kind of work done properly.

And this is where proofpurge comes into play: efficient, painless, and
extra-fast purge.

## Installation

This package can be installed with the go get command:

  go get github.com/joshakpeko/proofpurge

## Usage

One first needs to load debit-side and credit-side records into
separate csv files.

Default configuration suppose:

  - 1st column represents dates
  - 2nd column represents records label(containing reference string)
  - 3rd column represents amounts (transaction financial value)

Those default can be overriden by command-line arguments.

Note also that, except those 3 data needed by the program, additional
columns can contain additional data. The program neither cares nor
modifies theses additional data.

For basic usage:

  ./proofpure -df 'debit-file.csv' -cf 'credit-file.csv'

For more command-line options:

  ./proofpurge --help

The program outputs results in a new directory named 'purge'.
This directory contains purged debit-side and credit-side records in
separate csv file, and an additional log file containing details of
purged records.

## License

MIT License. (See LICENSE file)

## Author

Joshua Akpeko (2019)

email: jojoak@protonmail.ch
