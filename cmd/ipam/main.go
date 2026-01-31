package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/Shaman786/Bare-Metal-Post-Provisioning-Automation/internal/allocator"
	"github.com/Shaman786/Bare-Metal-Post-Provisioning-Automation/internal/config"
	"github.com/Shaman786/Bare-Metal-Post-Provisioning-Automation/internal/render"
	"github.com/Shaman786/Bare-Metal-Post-Provisioning-Automation/internal/report"
	"github.com/Shaman786/Bare-Metal-Post-Provisioning-Automation/internal/state"
	"github.com/Shaman786/Bare-Metal-Post-Provisioning-Automation/internal/validate"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: ipam <allocate|release> [ip]")
	}

	cfg := config.Load("config/ip-pools.yaml")
	pool := cfg.Pools[0]

	sm := state.Load("internal/state/assigned.json")
	defer sm.Close()

	st := sm.State

	switch os.Args[1] {

	case "allocate":
		ip := allocator.Allocate(pool, st.AssignedIPs)
		st.AssignedIPs = append(st.AssignedIPs, ip)
		sm.Save()

		json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"ip":      ip,
			"gateway": pool.Gateway,
			"dns":     pool.DNS,
		})

	case "release":
		if len(os.Args) != 3 {
			log.Fatal("usage: ipam release <ip>")
		}

		ip := os.Args[2]
		allocator.ValidateIPBelongsToPool(ip, pool)
		sm.Release(&st, ip)
		sm.Save()

		log.Printf("released IP %s\n", ip)

	case "cloud-init":
		ip := allocator.Allocate(pool, st.AssignedIPs)
		st.AssignedIPs = append(st.AssignedIPs, ip)
		sm.Save()

		prefix, _ := render.CIDRPrefix(pool.CIDR)

		cfg := render.NetworkConfig{
			Interface: "eth0",
			IP:        ip,
			Gateway:   pool.Gateway,
			DNS:       pool.DNS,
			CIDR:      prefix,
		}
		out, err := render.CloudInitNetwork(cfg)
		if err != nil {
			log.Fatal(err)
		}
		os.Stdout.Write([]byte(out))

	case "validate":
		const StatusReady = "READY"
		if len(os.Args) != 3 {
			log.Fatal("usage: ipam validate <ip>")
		}

		ip := os.Args[2]

		ok := validate.SSHReachable(ip)

		if !ok {
			log.Fatalf("validation failed for %s", ip)
		}
		if err := report.Write(
			"reports/"+ip+".json", ip, StatusReady,
		); err != nil {
			log.Fatalf("failed to write report for %s: %v", ip, err)
		}

	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}
