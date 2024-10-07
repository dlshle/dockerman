package gproxy

type Backend struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type ListenerCfg struct {
	Port     int        `yaml:"port"`
	Backends []*Backend `yaml:"backends"`
	Policy   string     `yaml:"policy"`
	Protocol string     `yaml:"protocol"`
}

type Config struct {
	Upstreams []*ListenerCfg `yaml:"upstreams"`
}
