// Package report handles the generation and persistence of provisioning reports
// (JSON status files) for tracking allocated resources.
package report

import (
	"encoding/json"
	"os"
	"time"
)

type DeliveryReport struct {
	IP        string    `json:"ip"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func Write(path string, ip string, status string) error {
	r := DeliveryReport{
		IP:        ip,
		Status:    status,
		Timestamp: time.Now(),
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(r)
}
