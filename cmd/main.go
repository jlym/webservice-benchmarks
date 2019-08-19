package main

import (
	"context"
	"fmt"

	"github.com/jlym/benchmark"
)

func main() {
	fmt.Println()

	b := benchmark.DataStore{}
	b.CreateTables(context.Background())
}
