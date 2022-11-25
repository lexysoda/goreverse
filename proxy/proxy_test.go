package proxy

import (
	"context"
	"math"
	"math/rand"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/docker/docker/api/types"
)

type mockCli struct {
	containers []types.Container
}

func (c *mockCli) ContainerList(ctx context.Context, opt types.ContainerListOptions) ([]types.Container, error) {
	return c.containers, nil
}

func TestRefreshHosts(t *testing.T) {

	cases := map[string][]string{
		"test":  []string{"https://google.de"},
		"test2": []string{},
		"test3": []string{"http://sub.dom.ain", "https://de.vu"},
	}

	for proxyLabel, containerLabels := range cases {
		m := &mockCli{}
		for _, label := range containerLabels {
			m.containers = append(
				m.containers,
				types.Container{
					Labels: map[string]string{proxyLabel: label},
					Ports: []types.Port{types.Port{
						PrivatePort: 80,
						PublicPort:  uint16(rand.Intn(math.MaxUint16 + 1)),
					}},
				},
			)
		}

		p := &Proxy{
			Hosts:    map[string]*httputil.ReverseProxy{},
			cli:      m,
			interval: 1,
			label:    proxyLabel,
		}
		p.refreshHosts()
		urlsGot := []string{}
		for hostname, _ := range p.Hosts {
			urlsGot = append(urlsGot, hostname)
		}
		urlsWant := []string{}
		for _, hostname := range containerLabels {
			u, err := url.Parse(hostname)
			if err != nil {
				t.Fatalf("Failed to parse: %s\n", hostname)
			}
			urlsWant = append(urlsWant, u.Hostname())
		}
		for i := 0; i < len(containerLabels); i++ {
			if urlsGot[i] != urlsWant[i] {
				t.Errorf("Got: %s, Want: %s\n", urlsGot[i], urlsWant[i])
			}
		}
	}
}
