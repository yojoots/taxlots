package main

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

// In weighted price calculations, we will consider floats with a difference less than this as "equal enough"
const FloatErrorTolerance = 0.000001

func TestReadTransactionLog(t *testing.T) {
	firstTransaction := "2021-01-01,buy,10000.00,1.00000000"
	secondTransaction := "2021-02-01,sell,20000.00,0.50000000"
	transactionLogReadResult := readTransactionLog(strings.NewReader("2021-01-01,buy,10000.00,1.00000000\n2021-02-01,sell,20000.00,0.50000000"))
	if transactionLogReadResult[0] != firstTransaction {
		t.Errorf("TransactionLog read error. Expected: \"%s\" ... got \"%s\" instead", firstTransaction, transactionLogReadResult[0])
	}
	if transactionLogReadResult[1] != secondTransaction {
		t.Errorf("TransactionLog read error. Expected: \"%s\" ... got \"%s\" instead", secondTransaction, transactionLogReadResult[1])
	}
}

func TestWeightedPrice(t *testing.T) {
	smallLotAtTenThousand := Lot{
		id:       1,
		date:     "2021-01-01",
		price:    10000.0,
		quantity: 1.00000000,
		txType:   "buy",
	}
	smallLotAtFiftyThousand := Lot{
		id:       1,
		date:     "2021-01-01",
		price:    50000.0,
		quantity: 1.00000000,
		txType:   "buy",
	}
	mediumLotAtTenThousand := Lot{
		id:       1,
		date:     "2021-01-01",
		price:    10000.0,
		quantity: 10.00000000,
		txType:   "buy",
	}
	mediumLotAtFortyThousand := Lot{
		id:       1,
		date:     "2021-01-01",
		price:    40000.0,
		quantity: 10.00000000,
		txType:   "buy",
	}
	largeLotAtTenThousand := Lot{
		id:       1,
		date:     "2021-01-01",
		price:    10000.0,
		quantity: 100.00000000,
		txType:   "buy",
	}
	largeLotAtTwentyThousand := Lot{
		id:       1,
		date:     "2021-01-01",
		price:    20000.0,
		quantity: 100.00000000,
		txType:   "buy",
	}

	expectedWeightS10M10 := float64(10000)
	if priceWeightResult := weightedPrice(smallLotAtTenThousand, mediumLotAtTenThousand); priceWeightResult != expectedWeightS10M10 {
		t.Errorf("Expected weighted price to be %f ... got %f instead", expectedWeightS10M10, priceWeightResult)
	}

	expectedWeightS10S50 := float64(30000)
	if priceWeightResult := weightedPrice(smallLotAtTenThousand, smallLotAtFiftyThousand); priceWeightResult != expectedWeightS10S50 {
		t.Errorf("Expected weighted price to be %f ... got %f instead", expectedWeightS10S50, priceWeightResult)
	}

	expectedWeightS10M40 := float64(37272.7272727)
	if priceWeightResult := weightedPrice(smallLotAtTenThousand, mediumLotAtFortyThousand); math.Abs(priceWeightResult-expectedWeightS10M40) > FloatErrorTolerance {
		t.Errorf("Expected weighted price to be %.10f ... got %.10f instead", expectedWeightS10M40, priceWeightResult)
	}

	expectedWeightS10L10 := float64(10000)
	if priceWeightResult := weightedPrice(smallLotAtTenThousand, largeLotAtTenThousand); priceWeightResult != expectedWeightS10L10 {
		t.Errorf("Expected weighted price to be %f ... got %f instead", expectedWeightS10L10, priceWeightResult)
	}

	expectedWeightM10L20 := float64(19090.9090909)
	if priceWeightResult := weightedPrice(mediumLotAtTenThousand, largeLotAtTwentyThousand); math.Abs(priceWeightResult-expectedWeightM10L20) > FloatErrorTolerance {
		t.Errorf("Expected weighted price to be %f ... got %f instead", expectedWeightM10L20, priceWeightResult)
	}

	expectedWeightL10L20 := float64(15000)
	if priceWeightResult := weightedPrice(largeLotAtTenThousand, largeLotAtTwentyThousand); priceWeightResult != expectedWeightL10L20 {
		t.Errorf("Expected weighted price to be %f ... got %f instead", expectedWeightL10L20, priceWeightResult)
	}

}
func TestParseRawTransaction(t *testing.T) {
	resultingLot, err := parseRawTransaction("2021-01-01,buy,10000.00,1.00000000", 0)
	if err != nil {
		t.Errorf(err.Error())
	}
	expectedLotResult := Lot{
		id:       1,
		date:     "2021-01-01",
		price:    10000.0,
		quantity: 1.00000000,
		txType:   "buy",
	}
	deepEqualResult := reflect.DeepEqual(resultingLot, expectedLotResult)
	if !deepEqualResult {
		t.Errorf("parseRawTransaction: Expected resultingLot to be %+s ... got %+s instead", expectedLotResult, resultingLot)
	}

	want := "1,2021-01-01,10000.00,1.00000000"
	if got := resultingLot.String(); got != want {
		t.Errorf("parseRawTransaction: Expected resultingLot.String() to be %s ... got %s instead", want, got)
	}
}
func TestSmallProcessTransactionsFIFO(t *testing.T) {
	resultingLots, err := processTransactions([]string{"2021-01-01,buy,10000.00,1.00000000", "2021-02-01,sell,20000.00,0.50000000"}, "fifo")
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(resultingLots) != 1 {
		t.Errorf("processTransactions: Expected 1 resulting lot back, got %d instead", len(resultingLots))
	}
	want := "1,2021-01-01,10000.00,0.50000000"
	if got := resultingLots[0].String(); got != want {
		t.Errorf("processTransactions: Expected resultingLots[0].String() to be %s ... got %s instead", want, got)
	}
}
func TestProcessTransactionsFIFO(t *testing.T) {
	resultingLots, err := processTransactions([]string{"2021-01-01,buy,10000.00,1.00000000", "2021-01-02,buy,20000.00,1.00000000", "2021-02-01,sell,20000.00,1.50000000"}, "fifo")
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(resultingLots) != 1 {
		t.Errorf("processTransactions: Expected 1 resulting lot back, got %d instead", len(resultingLots))
	}
	want := "2,2021-01-02,20000.00,0.50000000"
	if got := resultingLots[0].String(); got != want {
		t.Errorf("processTransactions: Expected resultingLots[0].String() to be %s ... got %s instead", want, got)
	}
}

func TestProcessTransactionsHIFO(t *testing.T) {
	resultingLots, err := processTransactions([]string{"2021-01-01,buy,10000.00,1.00000000", "2021-01-02,buy,20000.00,1.00000000", "2021-02-01,sell,20000.00,1.50000000"}, "hifo")
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(resultingLots) != 1 {
		t.Errorf("processTransactions: Expected 1 resulting lot back, got %d instead", len(resultingLots))
	}
	want := "1,2021-01-01,10000.00,0.50000000"
	if got := resultingLots[0].String(); got != want {
		t.Errorf("processTransactions: Expected resultingLots[0].String() to be %s ... got %s instead", want, got)
	}
}

func TestExcessiveSaleQuantity(t *testing.T) {
	resultingLots, err := processTransactions([]string{"2021-01-01,buy,10000.00,1.00000000", "2021-01-02,buy,20000.00,1.00000000", "2021-02-01,sell,20000.00,5.00000000"}, "hifo")
	if err == nil {
		t.Errorf("Sales exceeded buys, but no error resulted")
	}
	expectedErrorMessage := "Problem executing sale (hifo): Sale quantity exceeded total buy quantity; please ensure that transaction log input is valid"
	if err.Error() != expectedErrorMessage {
		t.Errorf("Unexpected error resulted from excessive sales. Expected: \"%s\" ... got \"%s\" instead", expectedErrorMessage, err.Error())
	}
	if len(resultingLots) > 0 {
		t.Errorf("Sales exceeded buys, but non-empty lot slice was returned")
	}
}

func TestBadAlgorithms(t *testing.T) {
	firstLots, err := processTransactions([]string{"2021-01-01,buy,10000.00,1.00000000", "2021-01-02,buy,20000.00,1.00000000", "2021-02-01,sell,20000.00,1.50000000"}, "lol")
	if err == nil {
		t.Errorf("Erroneous algorithm didn't elicit an error")
	}
	expectedErrorSnippet := "Invalid algorithm (must be either \"fifo\" or \"hifo\")"
	if !strings.Contains(err.Error(), expectedErrorSnippet) {
		t.Errorf("Unexpected error resulted from excessive sales. Expected: \"%s\" ... got \"%s\" instead", expectedErrorSnippet, err.Error())
	}
	if len(firstLots) > 0 {
		t.Errorf("Non-empty lot slice was returned despite erroneous algorithm")
	}

	secondLots, err := processTransactions([]string{"2021-01-01,buy,10000.00,1.00000000", "2021-01-02,buy,20000.00,1.00000000", "2021-02-01,sell,20000.00,1.50000000"}, "more than one word")
	if err == nil {
		t.Errorf("Erroneous algorithm didn't elicit an error (second attempt, multi-word)")
	}
	if !strings.Contains(err.Error(), expectedErrorSnippet) {
		t.Errorf("Unexpected error resulted from excessive sales. Expected: \"%s\" ... got \"%s\" instead", expectedErrorSnippet, err.Error())
	}
	if len(secondLots) > 0 {
		t.Errorf("Non-empty lot slice was returned despite erroneous algorithm")
	}
}

func TestBadInputs(t *testing.T) {
	extraFieldResult, err := processTransactions([]string{"2021-01-01,extraneousField,buy,10000.00,1.00000000", "2021-01-02,buy,20000.00,1.00000000", "2021-02-01,bad,20000.00,1.50000000"}, "fifo")
	if err == nil {
		t.Errorf("Extra nonsensical field didn't elicit an error")
	}
	expectedErrorSnippet := "Invalid tx format; incorrect argument count (should be 4, got 5)"
	if !strings.Contains(err.Error(), expectedErrorSnippet) {
		t.Errorf("Unexpected error resulted from bad txType. Expected: \"%s\" ... got \"%s\" instead", expectedErrorSnippet, err.Error())
	}
	if len(extraFieldResult) > 0 {
		t.Errorf("Non-empty lot slice was returned despite presence of extra field")
	}

	badTxTypeResult, err := processTransactions([]string{"2021-01-01,buy,10000.00,1.00000000", "2021-01-02,buy,20000.00,1.00000000", "2021-02-01,bad,20000.00,1.50000000"}, "fifo")
	if err == nil {
		t.Errorf("Erroneous txType didn't elicit an error")
	}
	expectedErrorSnippet = "Invalid order type (must be either \"buy\" or \"sell\")"
	if !strings.Contains(err.Error(), expectedErrorSnippet) {
		t.Errorf("Unexpected error resulted from bad txType. Expected: \"%s\" ... got \"%s\" instead", expectedErrorSnippet, err.Error())
	}
	if len(badTxTypeResult) > 0 {
		t.Errorf("Non-empty lot slice was returned despite erroneous txType")
	}

	badPriceResult, err := processTransactions([]string{"2021-01-01,buy,10000.00,1.00000000", "2021-01-02,buy,200f00.00,1.00000000", "2021-02-01,sell,20000.00,1.50000000"}, "fifo")
	if err == nil {
		t.Errorf("Erroneous price value didn't elicit an error")
	}
	expectedErrorSnippet = "Invalid (non-float) price"
	if !strings.Contains(err.Error(), expectedErrorSnippet) {
		t.Errorf("Unexpected error resulted from bad txType. Expected: \"%s\" ... got \"%s\" instead", expectedErrorSnippet, err.Error())
	}
	if len(badPriceResult) > 0 {
		t.Errorf("Non-empty lot slice was returned despite erroneous price value")
	}

	badQuantityResult, err := processTransactions([]string{"2021-01-01,buy,10000.00,1.00000000", "2021-01-02,buy,20000.00,1.0000xyz0", "2021-02-01,sell,20000.00,1.50000000"}, "fifo")
	if err == nil {
		t.Errorf("Erroneous quantity value didn't elicit an error")
	}
	expectedErrorSnippet = "Invalid (non-float) quantity"
	if !strings.Contains(err.Error(), expectedErrorSnippet) {
		t.Errorf("Unexpected error resulted from bad txType. Expected: \"%s\" ... got \"%s\" instead", expectedErrorSnippet, err.Error())
	}
	if len(badQuantityResult) > 0 {
		t.Errorf("Non-empty lot slice was returned despite erroneous quantity value")
	}
}
func TestEndToEnd(t *testing.T) {
	testInputs := []string{
		"2021-01-01,buy,10000.00,1.00000000\n2021-02-01,sell,20000.00,0.50000000",
		"2021-01-01,buy,10000.00,1.00000000\n2021-01-02,buy,20000.00,1.00000000\n2021-02-01,sell,20000.00,1.50000000",
		"2021-01-01,buy,10000.00,1.00000000\n2021-01-02,buy,20000.00,1.00000000\n2021-02-01,sell,20000.00,1.50000000",
		"2021-01-01,buy,10000.00,1.00000000\n2021-01-01,buy,15000.00,1.00000000\n2021-02-01,sell,20000.00,1.50000000",
		"2021-01-01,buy,10000.00,1.00000000\n2021-01-02,buy,20000.00,1.00000000\n2021-02-01,sell,20000.00,1.00000000",
		"2021-01-01,buy,10000.00,1.00000000\n2021-01-01,buy,15000.00,1.00000000\n2021-02-01,sell,20000.00,1.50000000\n2021-02-01,buy,21000.00,1.25000000",
		"2022-03-16,buy,55432.00,0.50000000\n2022-04-22,buy,58432.60,0.02000000\n2022-09-15,buy,82000.13,0.04000000\n2022-09-15,buy,83000.13,0.04000000\n2022-09-15,sell,91255.10,0.04200000",
		"2022-03-16,buy,55432.00,0.50000000\n2022-04-22,buy,58432.60,0.02000000\n2022-09-15,buy,82000.13,0.04000000\n2022-09-15,buy,83000.13,0.04000000\n2022-09-15,sell,91255.10,0.04200000",
		"2022-03-16,buy,55432.00,0.50000000\n2022-04-22,buy,58432.60,0.02000000\n2022-09-15,buy,82000.13,0.04000000\n2022-09-15,buy,83000.13,0.04000000\n2022-09-15,sell,91255.10,0.04200000\n2023-01-18,sell,89255.11,0.54000000",
		"2025-01-01,buy,1025000.00,0.01000000\n2025-01-02,buy,1025000.00,0.02000000\n2025-08-05,sell,2200000.00,0.00250000",
	}
	testAlgorithms := []string{"fifo", "fifo", "hifo", "hifo", "hifo", "hifo", "fifo", "hifo", "hifo", "fifo"}
	expectedResults := [][]string{
		{"1,2021-01-01,10000.00,0.50000000"},
		{"2,2021-01-02,20000.00,0.50000000"},
		{"1,2021-01-01,10000.00,0.50000000"},
		{"1,2021-01-01,12500.00,0.50000000"},
		{"1,2021-01-01,10000.00,1.00000000"},
		{"1,2021-01-01,12500.00,0.50000000", "2,2021-02-01,21000.00,1.25000000"},
		{"1,2022-03-16,55432.00,0.45800000", "2,2022-04-22,58432.60,0.02000000", "3,2022-09-15,82500.13,0.08000000"},
		{"1,2022-03-16,55432.00,0.50000000", "2,2022-04-22,58432.60,0.02000000", "3,2022-09-15,82500.13,0.03800000"},
		{"1,2022-03-16,55432.00,0.01800000"},
		{"1,2025-01-01,1025000.00,0.00750000", "2,2025-01-02,1025000.00,0.02000000"},
	}

	for idx, testInput := range testInputs {
		// Read transaction log
		transactionLogReadResult := readTransactionLog(strings.NewReader(testInput))

		// Process transactions
		lots, err := processTransactions(transactionLogReadResult, testAlgorithms[idx])
		if err != nil {
			t.Errorf("End-to-end test #%d failed: %s", idx, err.Error())
		}
		if len(lots) != len(expectedResults[idx]) {
			t.Errorf("End-to-end test #%d should leave %d lot(s) remaining ... left %d instead", idx, len(expectedResults[idx]), len(lots))
		}
		for resultIdx, expectedLotResult := range expectedResults[idx] {
			if lots[resultIdx].String() != expectedLotResult {
				t.Errorf("End-to-end test #%d(%d) should produce result: \"%s\" ... produced \"%s\" instead", idx, resultIdx, expectedLotResult, lots[resultIdx].String())
			}
		}
	}
}
