package utils

import (
	"fmt"
	"iter"
	"net"
)

// LocalIps возвращает список локальных IP-адресов
func LocalIps() iter.Seq[net.IP] {
	return func(yield func(net.IP) bool) {
		addrs, _ := net.InterfaceAddrs()
		for _, address := range addrs {
			// check if the address is a loopback or multicast address
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				if ipnet.IP.IsLoopback() {
					continue
				}

				yield(ipnet.IP)
			}
		}
	}
}

// CheckIpsInSameSubnet Проверка вхождения подсети
func CheckIpsInSameSubnet(ip1 string, ip2 string) bool {

	ip1IP := net.ParseIP(ip1)
	ip2IP := net.ParseIP(ip2)

	if ip1IP == nil || ip2IP == nil {
		return false
	}

	// Получаем маску по умолчанию для IPv4, которая равна 32 (все биты)
	defaultMask := net.CIDRMask(24, 32)
	network1 := ip1IP.Mask(defaultMask)
	network2 := ip2IP.Mask(defaultMask)

	fmt.Println(network1.String())
	fmt.Println(network2.String())

	return network1.Equal(network2)
}
