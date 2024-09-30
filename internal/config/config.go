package config

import (
	yaml "gopkg.in/yaml.v2"
)

type Reader struct {
	yaml.IsZeroer
}

type Config struct {
	DockerHost string `yaml:"dockerHost"`
}

type PortConfig struct {
	Source int `yaml:"source"` // source port in the container
	// exposed *int `yaml:"exposed"` // exposed port in the host, if not set, port will not be exposed
}

type ReadinessCheckConfig struct {
	Port                        int    `yaml:"port"`
	Path                        string `yaml:"path"`
	Method                      string `yaml:"method"`
	CheckIntervalSeconds        *int   `yaml:"checkIntervalSeconds"`
	HealthCheckFailureThreshold *int   `yaml:"healthCheckFailureThreshold"`
}

type AppConfig struct {
	Name           string                `yaml:"name"`
	Image          string                `yaml:"image"`
	Replicas       int                   `yaml:"replicas"`
	Ports          []PortConfig          `yaml:"ports,flow"`
	ReadinessCheck *ReadinessCheckConfig `yaml:"readinessCheck"`
}
