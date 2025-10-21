package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net"
	"os"
)

const AuthHeader = "User-Jwt"

// LocalIps возвращает список локальных IP-адресов
func LocalIps() []net.IP {

	retAddrs := make([]net.IP, 0)

	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
		// check if the address is a loopback or multicast address
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			if ipnet.IP.IsLoopback() {
				continue
			}

			retAddrs = append(retAddrs, ipnet.IP)
		}
	}

	return retAddrs
}

// CheckIpsInSameSubnet Проверка вхождения подсети
func CheckIpsInSameSubnet(ip1IP net.IP, ip2IP net.IP) bool {
	if ip1IP == nil || ip2IP == nil {
		return false
	}

	// Получаем маску по умолчанию для IPv4, которая равна 32 (все биты)
	defaultMask := net.CIDRMask(24, 32)
	network1 := ip1IP.Mask(defaultMask)
	network2 := ip2IP.Mask(defaultMask)

	return network1.Equal(network2)
}

func GetRequestJwt(c *gin.Context) (*jwt.Token, error) {
	tokenString := c.Request.Header.Get(AuthHeader)

	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
}
