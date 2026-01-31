package main

import (
	"encoding/json"
	"flag"
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
	// 1. Define Flags (Fixes hardcoded paths & single pool limitation)
	configPath := flag.String("config", "config/ip-pools.yaml", "Path to configuration file")
	statePath := flag.String("state", "internal/state/assigned.json", "Path to state file")
	poolName := flag.String("pool", "", "Specific pool name to use (defaults to first available)")

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("usage: ipam [flags] <allocate|release|cloud-init|validate> [ip]")
	}

	command := args[0]

	// 2. Load Config using Flags
	cfg := config.Load(*configPath)

	// 3. Logic to Select Pool (Fixes limitation where only pool[0] was used)
	var pool config.IPPool
	if *poolName == "" {
		if len(cfg.Pools) == 0 {
			log.Fatal("no IP pools defined in config")
		}
		pool = cfg.Pools[0]
	} else {
		found := false
		for _, p := range cfg.Pools {
			if p.Name == *poolName {
				pool = p
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("pool '%s' not found in configuration", *poolName)
		}
	}

	// 4. Load State using Flag
	sm := state.Load(*statePath)
	defer sm.Close()

	st := sm.State

	switch command {

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
		if len(args) != 2 {
			log.Fatal("usage: ipam release <ip>")
		}

		ip := args[1]
		allocator.ValidateIPBelongsToPool(ip, pool)
		sm.Release(&st, ip)
		sm.Save()

		log.Printf("released IP %s\n", ip)

	case "cloud-init":
		ip := allocator.Allocate(pool, st.AssignedIPs)
		st.AssignedIPs = append(st.AssignedIPs, ip)
		sm.Save()

		// Added error check here
		prefix, err := render.CIDRPrefix(pool.CIDR)
		if err != nil {
			log.Fatalf("invalid CIDR in config: %v", err)
		}

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
		if len(args) != 2 {
			log.Fatal("usage: ipam validate <ip>")
		}

		ip := args[1]

		ok := validate.SSHReachable(ip)

		if !ok {
			log.Fatalf("validation failed for %s", ip)
		}

		// Ensure reports directory exists before writing
		_ = os.MkdirAll("reports", 0o755)

		if err := report.Write(
			"reports/"+ip+".json", ip, StatusReady,
		); err != nil {
			log.Fatalf("failed to write report for %s: %v", ip, err)
		}

	default:
		log.Fatalf("unknown command: %s", command)
	}
}

// func main() {
// 	if len(os.Args) < 2 {
// 		log.Fatal("usage: ipam <allocate|release> [ip]")
// 	}

// 	cfg := config.Load("config/ip-pools.yaml")
// 	pool := cfg.Pools[0]

// 	sm := state.Load("internal/state/assigned.json")
// 	defer sm.Close()

// 	st := sm.State

// 	switch os.Args[1] {

// 	case "allocate":
// 		ip := allocator.Allocate(pool, st.AssignedIPs)
// 		st.AssignedIPs = append(st.AssignedIPs, ip)
// 		sm.Save()

// 		json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
// 			"ip":      ip,
// 			"gateway": pool.Gateway,
// 			"dns":     pool.DNS,
// 		})

// 	case "release":
// 		if len(os.Args) != 3 {
// 			log.Fatal("usage: ipam release <ip>")
// 		}

// 		ip := os.Args[2]
// 		allocator.ValidateIPBelongsToPool(ip, pool)
// 		sm.Release(&st, ip)
// 		sm.Save()

// 		log.Printf("released IP %s\n", ip)

// 	case "cloud-init":
// 		ip := allocator.Allocate(pool, st.AssignedIPs)
// 		st.AssignedIPs = append(st.AssignedIPs, ip)
// 		sm.Save()

// 		prefix, _ := render.CIDRPrefix(pool.CIDR)

// 		cfg := render.NetworkConfig{
// 			Interface: "eth0",
// 			IP:        ip,
// 			Gateway:   pool.Gateway,
// 			DNS:       pool.DNS,
// 			CIDR:      prefix,
// 		}
// 		out, err := render.CloudInitNetwork(cfg)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		os.Stdout.Write([]byte(out))

// 	case "validate":
// 		const StatusReady = "READY"
// 		if len(os.Args) != 3 {
// 			log.Fatal("usage: ipam validate <ip>")
// 		}

// 		ip := os.Args[2]

// 		ok := validate.SSHReachable(ip)

// 		if !ok {
// 			log.Fatalf("validation failed for %s", ip)
// 		}
// 		if err := report.Write(
// 			"reports/"+ip+".json", ip, StatusReady,
// 		); err != nil {
// 			log.Fatalf("failed to write report for %s: %v", ip, err)
// 		}

// 	default:
// 		log.Fatalf("unknown command: %s", os.Args[1])
// 	}
// }
