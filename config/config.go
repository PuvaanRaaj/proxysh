package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Domain struct {
	Domain string `yaml:"domain"`
	Target string `yaml:"target"`
	Auth   string `yaml:"auth,omitempty"`
	Log    string `yaml:"log,omitempty"` // full | minimal | off
}

type DaemonConfig struct {
	ListenPort int    `yaml:"listen_port"`
	LogFile    string `yaml:"log_file"`
	PIDFile    string `yaml:"pid_file"`
}

type CertConfig struct {
	CADir        string `yaml:"ca_dir"`
	CertDir      string `yaml:"cert_dir"`
	CACommonName string `yaml:"ca_common_name"`
	DaysValid    int    `yaml:"days_valid"`
}

type Config struct {
	Version int          `yaml:"version"`
	Daemon  DaemonConfig `yaml:"daemon"`
	Cert    CertConfig   `yaml:"cert"`
	Domains []Domain     `yaml:"domains"`
}

func Default() *Config {
	return &Config{
		Version: 1,
		Daemon: DaemonConfig{
			ListenPort: DaemonListenPort,
			LogFile:    DefaultLogFile(),
			PIDFile:    DefaultPIDFile(),
		},
		Cert: CertConfig{
			CADir:        DefaultCADir(),
			CertDir:      DefaultCertDir(),
			CACommonName: "devtun Local CA",
			DaysValid:    825,
		},
		Domains: []Domain{},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) AddDomain(domain, target string) {
	for i, d := range c.Domains {
		if d.Domain == domain {
			c.Domains[i].Target = target
			return
		}
	}
	c.Domains = append(c.Domains, Domain{Domain: domain, Target: target, Log: "minimal"})
}

func (c *Config) RemoveDomain(domain string) bool {
	for i, d := range c.Domains {
		if d.Domain == domain {
			c.Domains = append(c.Domains[:i], c.Domains[i+1:]...)
			return true
		}
	}
	return false
}

func FindConfigFile() string {
	dir, _ := os.Getwd()
	for {
		path := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ConfigFileName)
}
