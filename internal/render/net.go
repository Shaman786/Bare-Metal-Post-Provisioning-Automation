package render

import (
	"net"
	"strconv"
)

func CIDRPrefix(cidr string) (string, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	ones, _ := network.Mask.Size()
	return strconv.Itoa(ones), nil
}
