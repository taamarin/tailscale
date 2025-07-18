// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

//go:build linux

package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"tailscale.com/kube/kubetypes"
)

// Healthz is a simple health check server, if enabled it returns 200 OK if
// this tailscale node currently has at least one tailnet IP address else
// returns 503.
type Healthz struct {
	sync.Mutex
	hasAddrs bool
	podIPv4  string
}

func (h *Healthz) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Lock()
	defer h.Unlock()

	if h.hasAddrs {
		w.Header().Add(kubetypes.PodIPv4Header, h.podIPv4)
		if _, err := w.Write([]byte("ok")); err != nil {
			http.Error(w, fmt.Sprintf("error writing status: %v", err), http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "node currently has no tailscale IPs", http.StatusServiceUnavailable)
	}
}

func (h *Healthz) update(healthy bool) {
	h.Lock()
	defer h.Unlock()

	if h.hasAddrs != healthy {
		log.Println("Setting healthy", healthy)
	}
	h.hasAddrs = healthy
}

// registerHealthHandlers registers a simple health handler at /healthz.
// A containerized tailscale instance is considered healthy if
// it has at least one tailnet IP address.
func registerHealthHandlers(mux *http.ServeMux, podIPv4 string) *Healthz {
	h := &Healthz{podIPv4: podIPv4}
	mux.Handle("GET /healthz", h)
	return h
}
