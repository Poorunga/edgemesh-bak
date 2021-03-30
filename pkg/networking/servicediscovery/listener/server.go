package listener

import (
	"k8s.io/klog/v2"
)

// AddServer adds a server
func AddServer(svcName, svcPorts string) {
	ip := svcDesc.getIP(svcName)
	if ip != "" {
		svcDesc.set(svcName, ip, svcPorts)
		return
	}
	if len(unused) == 0 {
		// try to expand
		expandPool()
		if len(unused) == 0 {
			klog.Warningf("[EdgeMesh] insufficient fake IP !!")
			return
		}
	}
	ip = unused[0]
	unused = unused[1:]

	svcDesc.set(svcName, ip, svcPorts)
	err := dbmClient.Listener().Add(svcName, ip)
	if err != nil {
		klog.Errorf("[EdgeMesh] add listener %s to edge db error: %v", svcName, err)
		return
	}
}

// UpdateServer updates a server
func UpdateServer(svcName, svcPorts string) {
	ip := svcDesc.getIP(svcName)
	if ip == "" {
		if len(unused) == 0 {
			// try to expand
			expandPool()
			if len(unused) == 0 {
				klog.Warningf("[EdgeMesh] insufficient fake IP !!")
				return
			}
		}
		ip = unused[0]
		unused = unused[1:]
		err := dbmClient.Listener().Add(svcName, ip)
		if err != nil {
			klog.Errorf("[EdgeMesh] add listener %s to edge db error: %v", svcName, err)
		}
	}
	svcDesc.set(svcName, ip, svcPorts)
}


// DelServer deletes a server
func DelServer(svcName string) {
	ip := svcDesc.getIP(svcName)
	if ip == "" {
		return
	}
	svcDesc.del(svcName, ip)
	err := dbmClient.Listener().Del(svcName)
	if err != nil {
		klog.Errorf("[EdgeMesh] delete listener from edge db error: %v", err)
	}
	// recycling fakeIP
	unused = append(unused, ip)
}
