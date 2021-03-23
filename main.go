package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("p", 80, "Listening port")
	interval := flag.String("i", "60s", "Refresh interval as a go duration string")
	label := flag.String("l", "goreverse", "Container label")
	flag.Parse()

	p, err := New(*interval, *label)
	if err != nil {
		log.Fatal(err)
	}
	p.Start()

	http.Handle("/", p)

	url := fmt.Sprintf("localhost:%d", *port)
	log.Printf("Serving on %s\n", url)
	log.Fatal(http.ListenAndServe(url, nil))
}
