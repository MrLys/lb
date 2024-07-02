package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"ljos.app/lb/config"
)

var hosts sync.Map
var hostsList []string

type Host struct {
	Host     string
	Status   bool
	RevProxy *httputil.ReverseProxy
}

func main() {
	config := &config.Config{}
	config.Init()
	hostsList = config.Hosts
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	var lastIdx atomic.Uint32
	for _, host := range config.Hosts {
		url, err := url.Parse(host)
		if err != nil {
			log.Error("%q", err)
			continue
		}
		backend := Host{
			Host:     host,
			Status:   false,
			RevProxy: httputil.NewSingleHostReverseProxy(url),
		}
		hosts.Store(host, backend)
		go monitor(host, config.Endpoint, client)
	}
	go monitorMonitor(config.Hosts)
	// set up a http server to load balance
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		idx := int(lastIdx.Load())
		idx, host, err := roundRobin(idx)
		lastIdx.Store(uint32(idx))
		if err != nil {
			http.Error(w, "no healthy host", http.StatusServiceUnavailable)
			return
		}
		backend, ok := hosts.Load(host)
		if !ok {
			http.Error(w, "no healthy host", http.StatusServiceUnavailable)
			return
		}

		log.Info(fmt.Sprintf("host %s is healthy, revproxying %q\n", host, r.URL.Path))

		backend.(Host).RevProxy.ServeHTTP(w, r)
	})
	http.ListenAndServe(":8080", nil)
}
func IncrementCheckBounds(idx int) int {
	idx++
	if idx >= len(hostsList) {
		return 0
	}
	return idx
}
func roundRobin(idx int) (int, string, error) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	startIdx := idx

	idx = IncrementCheckBounds(idx)
	for {
		if idx == startIdx {
			return 0, "", errors.New("no healthy host" + fmt.Sprintf("%d", idx))
		}
		val, ok := hosts.Load(hostsList[idx])
		if !ok {
			log.Info(fmt.Sprintf("host not found %s\n", hostsList[idx]))
			idx = IncrementCheckBounds(idx)
			continue
		}

		if val.(Host).Status {
			return idx, hostsList[idx], nil
		}
		idx = IncrementCheckBounds(idx)
	}
}
func monitor(host string, endpoint string, client *http.Client) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	healthCheckURL := fmt.Sprintf("%s%s", host, endpoint)
	for {
		resp, err := client.Get(healthCheckURL)
		val, ok := hosts.Load(host)
		if !ok {
			log.Info(fmt.Sprintf("host not found %s\n", host))
			continue
		}
		v := val.(Host)
		if err != nil {
			log.Error("%q", err)
			v.Status = false
			hosts.Store(host, val)
		} else if resp.StatusCode == http.StatusOK {
			// fmt.Println("status code is 200 " + host + " is healthy")
			v.Status = true
			hosts.Store(host, v)
		} else {
			// fmt.Println("status code is not 200 " + host + " is not healthy")
			v.Status = false
			hosts.Store(host, v)
		}
		time.Sleep(5 * time.Second)
	}
}
func monitorMonitor(hostsIPs []string) {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	for {
		for _, host := range hostsIPs {
			val, ok := hosts.Load(host)
			if !ok {
				log.Info(fmt.Sprintf("host not found %s\n", host))
				continue
			}
			if val.(Host).Status {
				log.Info(fmt.Sprintf("host %s is healthy\n", host))
			} else {
				log.Info(fmt.Sprintf("host %s is not healthy\n", host))
			}

		}
		time.Sleep(6 * time.Second)
	}

}
