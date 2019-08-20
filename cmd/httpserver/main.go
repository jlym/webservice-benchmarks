package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/time", serveTime)
	mux.HandleFunc("/prime", servePrime)

	s := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Println("Serving")
	s.ListenAndServe()
}

func serveTime(w http.ResponseWriter, r *http.Request) {
	body := time.Now().String()
	w.Write([]byte(body))
}

func servePrime(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	nStr := values.Get("n")
	if nStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expected query param n"))
		return
	}

	n64, err := strconv.ParseInt(nStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("param n should be an int"))
		return
	}

	n := int(n64)

	body := fmt.Sprintf("%d", calcNthPrime(n))
	w.Write([]byte(body))
}

func calcNthPrime(n int) int {
	i := 2
	primeCount := 0
	for {
		if isPrime(i) {
			log.Println(i)
			primeCount++
			if primeCount == n {
				return i
			}
		}
		i++
	}
}

func isPrime(n int) bool {
	if n < 2 {
		return false
	}

	for i := 2; i < n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}
