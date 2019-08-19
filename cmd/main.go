package main

import (
	"fmt"

	"github.com/jlym/benchmark"
)

func main() {
	fmt.Println()

	b := benchmark.DataStore{}
	b.CreateTables()
}
