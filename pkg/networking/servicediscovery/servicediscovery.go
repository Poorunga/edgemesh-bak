package servicediscovery

import (
	"github.com/kubeedge/edgemesh/pkg/networking/servicediscovery/dns"
	"github.com/kubeedge/edgemesh/pkg/networking/servicediscovery/listener"
	"github.com/kubeedge/edgemesh/pkg/networking/servicediscovery/proxier"
)

// Init init
func Init() {
	// init tcp listener
	listener.Init()
	// init iptables
	proxier.Init()
	// init dns server
	dns.Init()
}

// Start starts all service discovery components
func Start() {
	go listener.StartListener()
	go proxier.StartProxier()
	go dns.StartDNS()
}

// Stop stop
func Stop() {
	proxier.Clean()
}
