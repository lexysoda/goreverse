package main

import (
	"log"
	"net/http"
)

func main() {
	p, err := New()
	if err != nil {
		log.Fatal(err)
	}
	p.Start()

	http.Handle("/", p)
	log.Fatal(http.ListenAndServe("localhost:80", nil))
}
