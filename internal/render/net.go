package render

import (
	"net"
)

func CIDRPrefix(cidr string) (string, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	ones, _ := network.Mask.Size()
	return string(rune(ones)), nil
}
