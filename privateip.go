package snowflake

import (
	"errors"
	"net"
)

// PrivateIPToMachineID convert private ip to machine id.
// Snowflake machineID max length is 6, max value is 63, use ip[3]
// From https://github.com/sony/sonyflake/blob/master/sonyflake.go
func PrivateIPToMachineID() int {
	ip, err := privateIPv4()
	if err != nil {
		return 0
	}
	return int(ip[3])
}

//--------------------------------------------------------------------
// private function defined.
//--------------------------------------------------------------------

func privateIPv4() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}

	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168)
}
