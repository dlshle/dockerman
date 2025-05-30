package dockerx

import "github.com/docker/docker/api/types/container"

type Image struct {
	ID       string
	RepoTags []string
	Created  int64
	Size     int64
}

type Container struct {
	ID            string
	Names         []string
	Image         string
	ImageID       string
	Labels        map[string]string
	State         string
	Status        string
	IPAddresses   map[string]string // network name to ip mapping
	ExposedPorts  map[uint16]uint16 // private port to public port
	RestartPolicy string
}

type Network struct {
	ID       string
	Name     string
	Driver   string
	Internal bool // is this network only for internal?
	Labels   map[string]string
	Peers    map[string]string // peer name and ip mapping
}

type RunOptions struct {
	RestartPolicy container.RestartPolicyMode
	ContainerName string
	Image         string
	Detached      bool
	VolumeMapping []string
	PortMapping   map[string]string // all tcp
	Envs          []string
	Networks      []string
	Labels        map[string]string
}
