package config

import (
	"github.com/baetyl/baetyl-core/config"
	"time"

	"github.com/baetyl/baetyl-go/http"
)

// Config the core config
type Config struct {
	config.Config
	Init InitConfig `yaml:"init" json:"init"`
}

type InitConfig struct {
	Batch struct {
		Name         string `yaml:"name" json:"name"`
		Namespace    string `yaml:"namespace" json:"namespace"`
		SecurityType string `yaml:"securityType" json:"securityType"`
		SecurityKey  string `yaml:"securityKey" json:"securityKey"`
	} `yaml:"batch" json:"batch"`
	Cloud struct {
		HTTP   http.ClientConfig `yaml:"http" json:"http"`
		Active struct {
			URL      string        `yaml:"url" json:"url" default:"/v1/active"`
			Interval time.Duration `yaml:"interval" json:"interval" default:"45s"`
		} `yaml:"active" json:"active"`
	} `yaml:"cloud" json:"cloud"`
	Edge struct {
		DownloadPath string `yaml:"downloadPath" json:"downloadPath" default:"var/lib/baetyl/download"`
	} `yaml:"edge" json:"edge"`
	ActivateConfig struct {
		Fingerprints []Fingerprint `yaml:"fingerprints" json:"fingerprints"`
		Attributes   []Attribute   `yaml:"attributes" json:"attributes"`
		Server       Server        `yaml:"server" json:"server"`
	} `yaml:"active" json:"active"`
}

// Server manually activated server configuration
type Server struct {
	Listen string `yaml:"listen" json:"listen"`
	Pages  string `yaml:"pages" json:"pages" default:"etc/baetyl/pages"`
}

// Fingerprint type to be collected
type Fingerprint struct {
	Proof Proof  `yaml:"proof" json:"proof"`
	Value string `yaml:"value" json:"value"`
}

// Attribute field to be filled
type Attribute struct {
	Name  string `yaml:"name" json:"name" validate:"nonzero"`
	Label string `yaml:"label" json:"label" validate:"nonzero"`
	Value string `yaml:"value" json:"value"`
	Desc  string `yaml:"description" json:"description"`
}

// Proof the proof of fingerprints
type Proof string

// all proofs
const (
	ProofSN         Proof = "sn"
	ProofInput      Proof = "input"
	ProofHostName   Proof = "hostName"
	ProofBootID     Proof = "bootID"
	ProofMachineID  Proof = "machineID"
	ProofSystemUUID Proof = "systemUUID"
)
