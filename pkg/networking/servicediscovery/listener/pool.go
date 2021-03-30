package listener

import (
	"strconv"

	"github.com/kubeedge/edgemesh/pkg/networking/servicediscovery/config"
)

var (
	unused      []string
	indexOfPool uint16
)

func initPool() {
	unused = make([]string, 0)
	// avoid 0.0
	indexOfPool = uint16(1)
	for ; indexOfPool <= uint16(255); indexOfPool++ {
		ip := config.Config.NetworkPrefix + getSubNet(indexOfPool)
		unused = append(unused, ip)
	}
}

// getSubNet converts uint16 to "uint8.uint8"
func getSubNet(subNet uint16) string {
	arg1 := uint64(subNet & 0x00ff)
	arg2 := uint64((subNet & 0xff00) >> 8)
	return strconv.FormatUint(arg2, 10) + "." + strconv.FormatUint(arg1, 10)
}

// expandPool expands fakeIP pool, each time with size of 256
func expandPool() {
	end := indexOfPool + uint16(255)
	for ; indexOfPool <= end; indexOfPool++ {
		// avoid 255.255
		if indexOfPool > config.Config.MaxPoolSize {
			return
		}
		ip := config.Config.NetworkPrefix + getSubNet(indexOfPool)
		// if ip is not used, append it to unused
		if svcDesc.getSvcPorts(ip) == "" {
			unused = append(unused, ip)
		}
	}
}
