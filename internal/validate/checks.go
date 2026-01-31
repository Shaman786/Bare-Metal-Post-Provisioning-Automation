package validate

import (
	"net"
	"time"
)

func SSHReachable(ip string) bool {
	conn, err := net.DialTimeout("tcp", ip+":22", 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
