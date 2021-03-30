package listener

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	dbmclient "github.com/kubeedge/edgemesh/pkg/common/dao/client"
	"github.com/kubeedge/edgemesh/pkg/controller"
	"github.com/kubeedge/edgemesh/pkg/networking/servicediscovery/config"
	"github.com/kubeedge/edgemesh/pkg/networking/trafficplugin/protocol"
	"github.com/kubeedge/edgemesh/pkg/networking/trafficplugin/protocol/http"
	"github.com/kubeedge/edgemesh/pkg/networking/trafficplugin/protocol/tcp"
)

const (
	SoOriginalDst = 80
)

var (
	dbmClient dbmclient.CoreInterface
	once      sync.Once
)

type sockAddr struct {
	family uint16
	data   [14]byte
}

func Init() {
	once.Do(func() {
		svcDesc = newServiceDescription()
		// init dao client
		dbmClient = dbmclient.NewDBMClient()
		// init fakeIP pool
		initPool()
		// recover service discovery meta from edge database and k8s database
		recoverFromDB()
	})
}

func StartListener() {
	for {
		conn, err := config.Config.Listener.Accept()
		if err != nil {
			klog.Warningf("[EdgeMesh] get tcp conn error: %v", err)
			continue
		}
		ip, port, err := realServerAddress(&conn)
		if err != nil {
			klog.Warningf("[EdgeMesh] get real destination of tcp conn error: %v", err)
			conn.Close()
			continue
		}
		proto, err := newProtocolFromSock(ip, port, conn)
		if err != nil {
			klog.Warningf("[EdgeMesh] get protocol from sock err: %v", err)
			conn.Close()
			continue
		}
		go proto.Process()
	}
}

// realServerAddress returns an intercepted connection's original destination.
func realServerAddress(conn *net.Conn) (string, int, error) {
	tcpConn, ok := (*conn).(*net.TCPConn)
	if !ok {
		return "", -1, fmt.Errorf("not a TCPConn")
	}

	file, err := tcpConn.File()
	if err != nil {
		return "", -1, err
	}

	// To avoid potential problems from making the socket non-blocking.
	tcpConn.Close()
	*conn, err = net.FileConn(file)
	if err != nil {
		return "", -1, err
	}

	defer file.Close()
	fd := file.Fd()

	var addr sockAddr
	size := uint32(unsafe.Sizeof(addr))
	err = getSockOpt(int(fd), syscall.SOL_IP, SoOriginalDst, uintptr(unsafe.Pointer(&addr)), &size)
	if err != nil {
		return "", -1, err
	}

	var ip net.IP
	switch addr.family {
	case syscall.AF_INET:
		ip = addr.data[2:6]
	default:
		return "", -1, fmt.Errorf("unrecognized address family")
	}

	port := int(addr.data[0])<<8 + int(addr.data[1])
	if err := syscall.SetNonblock(int(fd), true); err != nil {
		return "", -1, nil
	}

	return ip.String(), port, nil
}

func getSockOpt(s int, level int, name int, val uintptr, vallen *uint32) (err error) {
	_, _, e1 := syscall.Syscall6(syscall.SYS_GETSOCKOPT, uintptr(s), uintptr(level), uintptr(name), uintptr(val), uintptr(unsafe.Pointer(vallen)), 0)
	if e1 != 0 {
		err = e1
	}
	return
}

// newProtocolFromSock returns a protocol.Protocol interface if the ip is in proxier list
func newProtocolFromSock(ip string, port int, conn net.Conn) (proto protocol.Protocol, err error) {
	svcPorts := svcDesc.getSvcPorts(ip)
	protoName, svcName := getProtocol(svcPorts, port)
	if protoName == "" || svcName == "" {
		return nil, fmt.Errorf("protocol name: %s or svcName: %s is invalid", protoName, svcName)
	}

	svcNameSets := strings.Split(svcName, ".")
	if len(svcNameSets) != 2 {
		return nil, fmt.Errorf("invalid length %d after splitting svc name %s", len(svcNameSets), svcName)
	}
	namespace := svcNameSets[0]
	name := svcNameSets[1]

	switch protoName {
	case "http":
		proto = &http.HTTP{
			Conn:         conn,
			SvcName:      name,
			SvcNamespace: namespace,
			Port:         port,
		}
		err = nil
	case "tcp":
		proto = &tcp.TCP{
			Conn:         conn,
			SvcName:      name,
			SvcNamespace: namespace,
			Port:         port,
		}
		err = nil
	default:
		proto = nil
		err = fmt.Errorf("protocol: %s is not supported yet", protoName)
	}
	return
}

// getProtocol gets protocol name
func getProtocol(svcPorts string, port int) (string, string) {
	var protoName string
	sub := strings.Split(svcPorts, "|")
	n := len(sub)
	if n < 2 {
		return "", ""
	}
	svcName := sub[n-1]

	pstr := strconv.Itoa(port)
	if pstr == "" {
		return "", ""
	}
	for _, s := range sub {
		if strings.Contains(s, pstr) {
			protoName = strings.Split(s, ",")[0]
			break
		}
	}
	return protoName, svcName
}

// recoverFromDB gets fakeIP from edge database and assigns them to services after EdgeMesh starts
func recoverFromDB() {
	svcs, err := controller.GetServiceLister().Services(v1.NamespaceAll).List(labels.Everything())
	if err != nil || svcs == nil {
		klog.Errorf("[EdgeMesh] list all services from edge database error: %v", err)
		return
	}
	for _, svc := range svcs {
		svcName := svc.Namespace + "." + svc.Name
		ip, err := dbmClient.Listener().Get(svcName)
		if err != nil {
			klog.Errorf("[EdgeMesh] get listener of svc %s from edge database error: %v", svcName, err)
			continue
		}
		ip = strings.Trim(ip, "\"")
		if len(ip) == 0 {
			svcPorts := GetSvcPorts(svc, svcName)
			AddServer(svcName, svcPorts)
			klog.Warningf("[EdgeMesh] listener %s from edge database with no ip", svcName)
			continue
		}
		svcPorts := GetSvcPorts(svc, svcName)
		reserveIP(ip)
		svcDesc.set(svcName, ip, svcPorts)
		klog.Infof("[EdgeMesh] get fake ip `%s` --> %s", ip, svcName)
	}
}

func GetSvcPorts(svc *v1.Service, svcName string) string {
	svcPorts := ""
	for _, p := range svc.Spec.Ports {
		pro := strings.Split(p.Name, "-")
		sub := fmt.Sprintf("%s,%d,%d|", pro[0], p.Port, p.TargetPort.IntVal)
		svcPorts = svcPorts + sub
	}
	svcPorts += svcName
	return svcPorts
}

// reserveIp reserves used fakeIP
func reserveIP(ip string) {
	for i, value := range unused {
		if ip == value {
			unused = append(unused[:i], unused[i+1:]...)
			break
		}
	}
}
