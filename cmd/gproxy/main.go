package main

import (
	"flag"
	"log"
	"os"

	"github.com/dlshle/dockman/pkg/gproxy"
	"gopkg.in/yaml.v2"
)

func main() {
	var (
		cfgPath   string
		cfgString string
		err       error
	)
	flag.StringVar(&cfgPath, "config", "dmproxy.yaml", "path to the gproxy (yaml) config file")
	flag.StringVar(&cfgString, "data", "", "config data in yaml format")
	flag.Parse()

	if cfgPath == "" && cfgString == "" {
		log.Fatal("config path(config) or config data(data) is required")
	}

	var cfg *gproxy.Config

	if cfgPath != "" {
		cfg, err = loadConfig(cfgPath)
		if err != nil {
			log.Fatal(err)
		}
	} else if cfgString != "" {
		cfg, err = parseConfig([]byte(cfgString))
		if err != nil {
			log.Fatal(err)
		}
	}
	_, err = gproxy.Entry(cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func loadConfig(cfgPath string) (*gproxy.Config, error) {
	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	return parseConfig(cfgData)
}

func parseConfig(cfgData []byte) (*gproxy.Config, error) {
	cfg := &gproxy.Config{}
	if err := yaml.UnmarshalStrict(cfgData, cfg); err != nil {
		return nil, err
	}
	return cfg, nil

}
