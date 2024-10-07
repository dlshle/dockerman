package gproxy

import (
	"strconv"

	"gopkg.in/yaml.v2"
)

type Backend struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (b *Backend) Addr() string {
	return b.Host + ":" + strconv.Itoa(b.Port)
}

type ListenerCfg struct {
	Port     int        `yaml:"port"`
	Backends []*Backend `yaml:"backends,flow"`
	Policy   string     `yaml:"policy"`
	Protocol string     `yaml:"protocol"`
}

type Config struct {
	Upstreams []*ListenerCfg `yaml:"upstreams,flow"`
}

func UnmarshalConfig(data []byte) (*Config, error) {
	cfg := &Config{}
	err := yaml.UnmarshalStrict(data, &cfg)
	return cfg, err
}

func MarshalConfig(cfg *Config) ([]byte, error) {
	return yaml.Marshal(cfg)
}
