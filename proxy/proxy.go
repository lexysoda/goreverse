package proxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type dockerCli interface {
	ContainerList(context.Context, types.ContainerListOptions) ([]types.Container, error)
}

type Proxy struct {
	Hosts    map[string]*httputil.ReverseProxy
	cli      dockerCli
	interval time.Duration
	label    string
}

func New(interval string, label string) (*Proxy, error) {
	dur, err := time.ParseDuration(interval)
	if err != nil {
		return nil, err
	} else if dur <= 0 {
		return nil, fmt.Errorf("Duration must be > 0")
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Proxy{
		Hosts:    map[string]*httputil.ReverseProxy{},
		cli:      cli,
		interval: dur,
		label:    label,
	}, nil
}

func (p *Proxy) Start() {
	t := time.NewTicker(p.interval)
	go func() {
		for {
			p.refreshHosts()
			<-t.C
		}
	}()
}

func (p *Proxy) refreshHosts() {
	log.Println("Refreshing host mappings")
	newHosts := map[string]*httputil.ReverseProxy{}

	args := filters.NewArgs()
	args.Add("label", p.label)
	containers, err := p.cli.ContainerList(context.Background(), types.ContainerListOptions{Filters: args})
	if err != nil {
		log.Printf("Failed to query the docker api: %s\n", err)
		return
	}

	for _, c := range containers {
		from, err := url.Parse(c.Labels[p.label])
		if err != nil {
			log.Printf("Container %s contains invalid goreverse Url: %s\n", c.ID, c.Labels[p.label])
			continue
		}
		var to *url.URL
		for _, port := range c.Ports {
			if port.PrivatePort == 80 {
				to, _ = url.Parse(fmt.Sprintf("http://localhost:%d", port.PublicPort))
			}
		}
		if to == nil {
			continue
		}
		proxy, ok := p.Hosts[from.Hostname()]
		if ok {
			newHosts[from.Hostname()] = proxy
		} else {
			newHosts[from.Hostname()] = httputil.NewSingleHostReverseProxy(to)
		}
	}
	p.Hosts = newHosts
	log.Printf("Refreshed hosts: %v\n", p.Hosts)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	out := fmt.Sprintf("Received request for %s: ", r.Host)
	h, ok := p.Hosts[r.Host]
	if !ok {
		log.Printf("%sNo matching entry found\n", out)
		return
	}
	log.Printf("%sRedirecting to %s\n", out, r.Host)
	h.ServeHTTP(w, r)
}
