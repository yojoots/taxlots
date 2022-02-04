package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Lot struct {
	id       int
	date     string
	price    float64
	quantity float64
	txType   string
}

func (lot Lot) String() string {
	return fmt.Sprintf("%d,%s,%.2f,%.8f", lot.id, lot.date, lot.price, lot.quantity)
}

// Function to calculate weighted price of lot in cases where multiple buys occurred on the same date
func weightedPrice(oldLot Lot, newLot Lot) float64 {
	quantityTotal := oldLot.quantity + newLot.quantity
	oldLotWeight := oldLot.quantity / quantityTotal
	newLotWeight := newLot.quantity / quantityTotal
	return ((oldLot.price * oldLotWeight) + (newLot.price * newLotWeight))
}

// Function to print to stdout a descriptive error message and exit the script with a non-zero exit code
func errorAndExit(errorMsg string) {
	fmt.Printf("ERROR: %s\n\n", errorMsg)
	fmt.Printf("Example usage:\necho -e '2021-01-01,buy,10000.00,1.00000000\\n2021-02-01,sell,20000.00,0.50000000' | taxlots fifo\n")
	os.Exit(1)
}

// Function to execute a single sale transaction, subtracting saleQuantity from existing tax lots
// Note: this function assumes that the lots are sorted such that the head of the slice is prioritized
// which means that it is the responsibility of the calling function to sort lots before calling executeSale
func executeSale(lots []Lot, saleQuantity float64) ([]Lot, error) {
	for saleQuantity > 0 && len(lots) > 0 {
		if lots[0].quantity > saleQuantity {
			lots[0].quantity -= saleQuantity
			saleQuantity = 0
		} else if lots[0].quantity == saleQuantity {
			lots = lots[1:]
			saleQuantity = 0
		} else {
			// Reaching here means that lots[0].quantity < saleQuantity
			saleQuantity -= lots[0].quantity
			lots = lots[1:]
		}
	}
	if saleQuantity > 0 {
		// Reaching here means that input contained more sales than buys; interpret as erroneous
		return nil, fmt.Errorf("Sale quantity exceeded total buy quantity; please ensure that transaction log input is valid")
	}
	return lots, nil
}

// Function to parse a raw transaction string (in CSV format) into a Lot structure and txType (either "buy" or "sell")
func parseRawTransaction(rawTx string, lotCount int) (Lot, error) {
	txArray := strings.Split(rawTx, ",")
	if len(txArray) != 4 {
		return Lot{}, fmt.Errorf("Invalid tx format; incorrect argument count (should be 4, got %d): %s", len(txArray), rawTx)
	}

	txDate := txArray[0]
	txType := strings.ToLower(txArray[1])
	if txType != "buy" && txType != "sell" {
		return Lot{}, fmt.Errorf("Invalid order type (must be either \"buy\" or \"sell\"): %s", txType)
	}
	txPrice, err := strconv.ParseFloat(txArray[2], 64)
	if err != nil {
		return Lot{}, fmt.Errorf("Invalid (non-float) price: %s", txArray[2])
	}
	txQuantity, err := strconv.ParseFloat(txArray[3], 64)
	if err != nil {
		return Lot{}, fmt.Errorf("Invalid (non-float) quantity: %s", txArray[3])
	}

	lot := Lot{
		id:       lotCount + 1,
		date:     txDate,
		price:    txPrice,
		quantity: txQuantity,
		txType:   txType,
	}

	return lot, nil
}

// Function to process all transactions in a transaction log
// transactions must be an array of CSV strings representing the raw transaction details, in chronological order
// algorithm must be either "fifo" or "hifo"
// Returns remaining lots after processing is complete
func processTransactions(transactions []string, algorithm string) (lots []Lot, err error) {
	// First check to ensure algorithm is valid
	if algorithm != "fifo" && algorithm != "hifo" {
		return nil, fmt.Errorf("Invalid algorithm (must be either \"fifo\" or \"hifo\"): %s", algorithm)
	}

	// Loop through all transactions and process them in order
	for _, tx := range transactions {
		newLot, err := parseRawTransaction(tx, len(lots))
		if err != nil {
			return nil, fmt.Errorf("Problem parsing raw transaction (%s): %s", tx, err.Error())
		}
		switch newLot.txType {
		case "buy":
			if len(lots) == 0 || lots[len(lots)-1].date != newLot.date {
				// Buy lot with never-before-seen date
				lots = append(lots, newLot)
			} else {
				// Buys on same date are aggregated into a single lot with a weighted-average price
				lots[len(lots)-1].price = weightedPrice(lots[len(lots)-1], newLot)
				lots[len(lots)-1].quantity += newLot.quantity
			}
		case "sell":
			switch algorithm {
			case "fifo":
				// Subtract from lots[0]
				// (we're assuming our transactions list is in chronological order)
				lots, err = executeSale(lots, newLot.quantity)
				if err != nil {
					return nil, fmt.Errorf("Problem executing sale (fifo): %s", err.Error())
				}
			case "hifo":
				// Execute hifo on a sorted-by-price list of lots
				sort.SliceStable(lots, func(i, j int) bool {
					return lots[i].price > lots[j].price
				})
				// Now that lots is sorted in highest-price-first order, execute the sale
				lots, err = executeSale(lots, newLot.quantity)
				if err != nil {
					return nil, fmt.Errorf("Problem executing sale (hifo): %s", err.Error())
				}
				// After processing, sort lots back to default chronological ordering
				sort.SliceStable(lots, func(i, j int) bool {
					return lots[i].id < lots[j].id
				})
			default:
				return nil, fmt.Errorf("Invalid algorithm (must be either \"fifo\" or \"hifo\"): %s", algorithm)
			}
		default:
			return nil, fmt.Errorf("Invalid order type (must be either \"buy\" or \"sell\"): %s", newLot.txType)
		}
	}
	return
}

// Helper function to read transactionLog from stdin
func readTransactionLog(in io.Reader) (transactionLog []string) {
	scanner := bufio.NewScanner(in)
	for {
		scanner.Scan()
		// Store scanned string
		text := scanner.Text()
		if len(text) != 0 {
			transactionLog = append(transactionLog, text)
		} else {
			// exit on empty string
			break
		}
	}
	return
}

func main() {
	// Ensure that provided arguments are in expected format
	if len(os.Args) != 2 {
		errorAndExit("Must pass in chosen tax algorithm (\"fifo\" or \"hifo\") as first and only argument")
	}
	chosenAlgorithm := os.Args[1]
	if chosenAlgorithm != "fifo" && chosenAlgorithm != "hifo" {
		errorAndExit(fmt.Sprintf("Invalid algorithm (must be either \"fifo\" or \"hifo\"): %s", chosenAlgorithm))
	}

	// Read transactionLog from stdin
	transactionLog := readTransactionLog(os.Stdin)

	// Process transactions
	lots, err := processTransactions(transactionLog, chosenAlgorithm)
	if err != nil {
		errorAndExit(err.Error())
	}

	// Print results (remaining tax lots) after processing is complete, separated by newlines
	for _, lot := range lots {
		fmt.Printf("%s\n", lot.String())
	}
}
