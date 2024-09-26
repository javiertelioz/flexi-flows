package use_cases

import (
	"log"
)

// IterateFunc iterates over each item and processes it.
func IterateFunc(data any) (any, error) {
	item := data
	log.Printf(" item: %v", item)
	return item, nil
}
