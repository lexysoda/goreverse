package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Proxy struct {
	Hosts map[string]*url.URL
	cli   *client.Client
	sync.Mutex
}

func New() (*Proxy, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Proxy{
		Hosts: map[string]*url.URL{},
		cli:   cli,
	}, nil
}

func (p *Proxy) Start() {
	t := time.NewTicker(time.Minute)
	go func() {
		for {
			p.refreshHosts()
			<-t.C
		}
	}()
}

func (p *Proxy) refreshHosts() {
	log.Println("Refreshing host mappings")
	newHosts := map[string]*url.URL{}
	args := filters.NewArgs()
	args.Add("label", "goreverse")
	containers, err := p.cli.ContainerList(context.Background(), types.ContainerListOptions{Filters: args})
	if err != nil {
		log.Printf("Failed to query the docker api: %s\n", err)
		return
	}

	for _, c := range containers {
		from, err := url.Parse(c.Labels["goreverse"])
		if err != nil {
			log.Printf("Container %s contains invalid goreverse Url: %s\n", c.ID, c.Labels["goreverse"])
		}
		var to *url.URL
		for _, port := range c.Ports {
			if port.PrivatePort == 80 {
				to, _ = url.Parse(fmt.Sprintf("http://localhost:%d", port.PublicPort))
			}
		}
		newHosts[from.Hostname()] = to
	}

	p.Lock()
	p.Hosts = newHosts
	p.Unlock()
	log.Printf("Refreshed hosts: %v\n", p.Hosts)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	out := fmt.Sprintf("Received request for %s: ", r.Host)
	if h, ok := p.Hosts[r.Host]; ok {
		log.Printf("%sRedirecting to %s\n", out, h)
		proxy := httputil.NewSingleHostReverseProxy(h)
		proxy.ServeHTTP(w, r)
		return
	}
	log.Printf("%sNo matching entry found\n", out)
}
