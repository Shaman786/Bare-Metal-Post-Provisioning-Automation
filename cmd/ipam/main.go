package main

import (
	"encoding/json"
	"log"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

type IPPool struct {
	Name     string   `yaml:"name"`
	CIDR     string   `yaml:"cidr"`
	Gateway  string   `yaml:"gateway"`
	DNS      []string `yaml:"dns"`
	Reserved []string `yaml:"reserved"`
}

type IPPoolConfig struct {
	Pools []IPPool `yaml:"pools"`
}

type State struct {
	AssignedIPs []string `json:"assigned_ips"`
}

func main() {
	configFile := "config/ip-pools.yaml"
	stateFile := "internal/state/assigned.json"

	cfg := loadConfig(configFile)
	state := loadState(stateFile)

	pool := cfg.Pools[0] // MVP: single pool
	ip := allocateIP(pool, state)

	state.AssignedIPs = append(state.AssignedIPs, ip)
	saveState(stateFile, state)

	output := map[string]interface{}{
		"ip":      ip,
		"gateway": pool.Gateway,
		"dns":     pool.DNS,
	}

	json.NewEncoder(os.Stdout).Encode(output)
}

func loadConfig(path string) IPPoolConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	var cfg IPPoolConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatal(err)
	}
	return cfg
}

func loadState(path string) State {
	data, err := os.ReadFile(path)
	if err != nil {
		return State{}
	}

	var state State
	json.Unmarshal(data, &state)
	return state
}

func saveState(path string, state State) {
	data, _ := json.MarshalIndent(state, "", "  ")
	os.WriteFile(path, data, 0o644)
}

func allocateIP(pool IPPool, state State) string {
	_, network, err := net.ParseCIDR(pool.CIDR)
	if err != nil {
		log.Fatal(err)
	}

	reserved := make(map[string]bool)
	for _, ip := range pool.Reserved {
		reserved[ip] = true
	}
	for _, ip := range state.AssignedIPs {
		reserved[ip] = true
	}

	for ip := network.IP.Mask(network.Mask); network.Contains(ip); incIP(ip) {
		ipStr := ip.String()
		if !reserved[ipStr] {
			return ipStr
		}
	}

	log.Fatal("No available IPs")
	return ""
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
