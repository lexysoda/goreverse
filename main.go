package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/lexysoda/goreverse/proxy"
)

func main() {
	port := flag.Int("p", 80, "Listening port")
	interval := flag.String("i", "60s", "Refresh interval as a go duration string")
	label := flag.String("l", "goreverse", "Container label")
	certFile := flag.String("cert", "", "Path to certificate file")
	keyFile := flag.String("key", "", "Path to key file") 
	flag.Parse()

	p, err := proxy.New(*interval, *label)
	if err != nil {
		log.Fatal(err)
	}
	p.Start()

	http.Handle("/", p)

	url := fmt.Sprintf("localhost:%d", *port)
	log.Printf("Serving on %s\n", url)
	log.Fatal(http.ListenAndServeTLS(url, *certFile, *keyFile, nil))
}
