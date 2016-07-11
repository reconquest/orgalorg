package main

import (
	"fmt"
)

type bytesStringer struct {
	Amount int
}

func (stringer bytesStringer) String() string {
	amount := float64(stringer.Amount)

	suffixes := map[string]string{
		"b":   "KiB",
		"KiB": "MiB",
		"MiB": "GiB",
		"GiB": "TiB",
	}

	suffix := "b"
	for amount >= 1024 {
		if newSuffix, ok := suffixes[suffix]; ok {
			suffix = newSuffix
		} else {
			break
		}

		amount /= 1024
	}

	return fmt.Sprintf("%.2f%s", amount, suffix)
}
