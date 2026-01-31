package allocator

import (
	"bytes"
	"log"
	"net"
	"sort"

	"github.com/Shaman786/Bare-Metal-Post-Provisioning-Automation/internal/config"
)

func Allocate(pool config.IPPool, assigned []string) string {
	_, network, err := net.ParseCIDR(pool.CIDR)
	if err != nil {
		log.Fatal(err)
	}

	reserved := make(map[string]bool)

	for _, ip := range pool.Reserved {
		reserved[ip] = true
	}

	for _, ip := range assigned {
		reserved[ip] = true
	}
	var candidates []net.IP
	for ip := network.IP.Mask(network.Mask); network.Contains(ip); incIP(ip) {
		if isNetworkIP(ip, network) || isBroadcastIP(ip, network) {
			continue
		}
		ipStr := ip.String()
		if !reserved[ipStr] {
			candidates = append(candidates, append(net.IP(nil), ip...))
		}
	}
	if len(candidates) == 0 {

		log.Fatal("no available IPs")
	}
	sort.Slice(candidates, func(i, j int) bool {
		return bytes.Compare(candidates[i], candidates[j]) < 0
	})
	return candidates[0].String()
}

func ValidateIPInCIDR(ip string, cidr string) {
	parsedIP := net.ParseIP(ip)
	_, network, _ := net.ParseCIDR(cidr)

	if !network.Contains(parsedIP) {
		log.Fatalf("SECURITY VIOLATION: IP %s outside CIDR %s", ip, cidr)
	}
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func isNetworkIP(ip net.IP, network *net.IPNet) bool {
	return ip.Equal(network.IP)
}

func isBroadcastIP(ip net.IP, network *net.IPNet) bool {
	broadcast := make(net.IP, len(ip))
	for i := range ip {
		broadcast[i] = network.IP[i] | ^network.Mask[i]
	}
	return ip.Equal(broadcast)
}

func ValidateIPBelongsToPool(ip string, pool config.IPPool) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		log.Fatalf("invalid IP address: %s", ip)
	}

	_, network, err := net.ParseCIDR(pool.CIDR)
	if err != nil {
		log.Fatal(err)
	}

	if !network.Contains(parsed) {
		log.Fatalf("IP %s does not belong to pool %s", ip, pool.CIDR)
	}
}
