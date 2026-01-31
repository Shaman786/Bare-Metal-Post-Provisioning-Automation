// Package render handles the generation of configuration files (e.g., cloud-init)
// and network settings formatting.
package render

import (
	"bytes"
	"text/template"
)

type NetworkConfig struct {
	Interface string
	IP        string
	Gateway   string
	DNS       []string
	CIDR      string
}

const cloudInitTemplate = `#cloud-config
network:
  version: 2
  ethernets:
    {{ .Interface }}:
      dhcp4: false
      addresses:
        - {{ .IP }}/{{ .CIDR }}
      gateway4: {{ .Gateway }}
      nameservers:
        addresses: [{{ range $i, $v := .DNS }}{{ if $i }}, {{ end }}{{ $v }}{{ end }}]
`

func CloudInitNetwork(cfg NetworkConfig) (string, error) {
	tpl, err := template.New("cloudinit").Parse(cloudInitTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, cfg); err != nil {
		return "", err
	}

	return buf.String(), nil
}
