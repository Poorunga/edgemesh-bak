package listener

import (
	"sync"
)

var svcDesc *ServiceDescription

type ServiceDescription struct {
	sync.RWMutex
	SvcPortsByIP map[string]string // key: fakeIP, value: SvcPorts
	IPBySvc      map[string]string // key: svcName.svcNamespace, value: fakeIP
}

func newServiceDescription() *ServiceDescription {
	return &ServiceDescription{
		SvcPortsByIP: make(map[string]string),
		IPBySvc:      make(map[string]string),
	}
}

// set is a thread-safe operation to add to map
func (sd *ServiceDescription) set(svcName, ip, svcPorts string) {
	sd.Lock()
	defer sd.Unlock()
	sd.IPBySvc[svcName] = ip
	sd.SvcPortsByIP[ip] = svcPorts
}

// del is a thread-safe operation to del from map
func (sd *ServiceDescription) del(svcName, ip string) {
	sd.Lock()
	defer sd.Unlock()
	delete(sd.IPBySvc, svcName)
	delete(sd.SvcPortsByIP, ip)
}

// getIP is a thread-safe operation to get from map
func (sd *ServiceDescription) getIP(svcName string) string {
	sd.RLock()
	defer sd.RUnlock()
	ip := sd.IPBySvc[svcName]
	return ip
}

// getSvcPorts is a thread-safe operation to get from map
func (sd *ServiceDescription) getSvcPorts(ip string) string {
	sd.RLock()
	defer sd.RUnlock()
	svcPorts := sd.SvcPortsByIP[ip]
	return svcPorts
}

// GetServiceServer returns the proxier IP by given servicediscovery name
func GetServiceServer(svcName string) string {
	ip := svcDesc.getIP(svcName)
	return ip
}
