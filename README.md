# Tax Lot Processor

## Information
An important part of a brokerage product is keeping track of tax lots. A tax lot is created when a purchase is made. When a sale is made, the tax lots deducted by the sale are determined by a chosen algorithm. This processor parses a transaction log and outputs the remaining tax lots based on the chosen algorithm (`fifo` or `hifo`).

## Requirements

* A modern version of [Golang](https://go.dev/doc/install) must be installed

## Installation

```bash
git clone git@github.com:yojoots/taxlots.git
cd taxlots
go install
```

This will create a `taxlots` executable binary in your `$GOBIN` directory (by default [`$GOPATH/bin`](https://pkg.go.dev/cmd/go#hdr-GOPATH_environment_variable) or `$HOME/go/bin`) which can be executed on the command-line with the `taxlots` command.

Alternatively, you can use `go build` (instead of `go install`) to have a `taxlots` binary generated in your current working directory, and either call it with `./taxlots` or move it to a folder that is included in your `$PATH` environment variable to remove the need for the relative path prefix of `./`

### Implementation details

* The script takes one argument and reads a transaction log from stdin in the format of `date,buy/sell,price,quantity` separated by line breaks
* Transactions are expected to be provided in chronological order
* The argument passed into the script determines the tax lot selection algorithm
  * `fifo` - the first lots bought are the first lots sold
  * `hifo` - the first lots sold are the lots with the highest price
* Lots are tracked internally by an incrementing integer id starting at 1
  * Buys on the same date are aggregated into a single lot, the `price` is the weighted average price, the `id` remains the same
* After the transaction log is processed, the remaining lots (in the format of `id,date,price,quantity`) are printed to stdout
  * `price` shown with two decimal places
  * `quantity` shown with eight decimal places
* If an error is encountered, a descriptive error message is printed to stdout and the script exits with a non-zero exit code
* Automated tests are included in [`main_test.go`](main_test.go)

## Testing

Unit tests can be run with `go test` (or `go test -v` if you want verbose output)

## Example Usage
```bash
$ echo -e '2021-01-01,buy,10000.00,1.00000000\n2021-02-01,sell,20000.00,0.50000000' | taxlots fifo
1,2021-01-01,10000.00,0.50000000

$ echo -e '2021-01-01,buy,10000.00,1.00000000\n2021-01-02,buy,20000.00,1.00000000\n2021-02-01,sell,20000.00,1.50000000' | taxlots fifo
2,2021-01-02,20000.00,0.50000000

$ echo -e '2021-01-01,buy,10000.00,1.00000000\n2021-01-02,buy,20000.00,1.00000000\n2021-02-01,sell,20000.00,1.50000000' | taxlots hifo
1,2021-01-01,10000.00,0.50000000
```