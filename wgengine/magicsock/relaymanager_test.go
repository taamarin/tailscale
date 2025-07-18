// Copyright (c) Tailscale Inc & AUTHORS
// SPDX-License-Identifier: BSD-3-Clause

package magicsock

import (
	"net/netip"
	"testing"

	"tailscale.com/disco"
	"tailscale.com/types/key"
	"tailscale.com/util/set"
)

func TestRelayManagerInitAndIdle(t *testing.T) {
	rm := relayManager{}
	rm.startUDPRelayPathDiscoveryFor(&endpoint{}, netip.AddrPort{}, addrQuality{}, false)
	<-rm.runLoopStoppedCh

	rm = relayManager{}
	rm.stopWork(&endpoint{})
	<-rm.runLoopStoppedCh

	rm = relayManager{}
	rm.handleCallMeMaybeVia(&endpoint{c: &Conn{discoPrivate: key.NewDisco()}}, addrQuality{}, false, &disco.CallMeMaybeVia{UDPRelayEndpoint: disco.UDPRelayEndpoint{ServerDisco: key.NewDisco().Public()}})
	<-rm.runLoopStoppedCh

	rm = relayManager{}
	rm.handleRxDiscoMsg(&Conn{discoPrivate: key.NewDisco()}, &disco.BindUDPRelayEndpointChallenge{}, &discoInfo{}, epAddr{})
	<-rm.runLoopStoppedCh

	rm = relayManager{}
	rm.handleRelayServersSet(make(set.Set[candidatePeerRelay]))
	<-rm.runLoopStoppedCh

	rm = relayManager{}
	rm.getServers()
	<-rm.runLoopStoppedCh
}

func TestRelayManagerGetServers(t *testing.T) {
	rm := relayManager{}
	servers := make(set.Set[candidatePeerRelay], 1)
	c := candidatePeerRelay{
		nodeKey:  key.NewNode().Public(),
		discoKey: key.NewDisco().Public(),
	}
	servers.Add(c)
	rm.handleRelayServersSet(servers)
	got := rm.getServers()
	if !servers.Equal(got) {
		t.Errorf("got %v != want %v", got, servers)
	}
}
