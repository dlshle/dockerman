package netx

import (
	"fmt"
	"net"
)

func IsHostReachable(ip string, port int) bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	defer conn.Close()
	return err == nil
}
