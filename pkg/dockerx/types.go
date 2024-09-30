package dockerx

type Image struct {
	ID       string
	RepoTags []string
	Created  int64
	Size     int64
}

type Container struct {
	ID           string
	Names        []string
	Image        string
	ImageID      string
	Labels       map[string]string
	State        string
	Status       string
	IPAddresses  map[string]string // network name to ip mapping
	ExposedPorts []uint16
}

type Network struct {
	Name     string
	Driver   string
	Internal bool // is this network only for internal?
	Labels   map[string]string
	Peers    map[string]string // peer name and ip mapping
}

type RunOptions struct {
	ContainerName string
	Image         string
	Detached      bool
	VolumeMapping []string
	PortMapping   map[string]string
	Envs          []string
	Networks      []string
	Labels        map[string]string
}
