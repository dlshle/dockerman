package probe

const (
	ProbeTypeHTTP = "http"
	ProbeTypeTCP  = "tcp"
)

type HTTPProbeConfig struct {
	URL        string  `json:"url"`
	Method     string  `json:"method"`
	Base64Body *string `json:"base64Body"`
	Headers    []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"headers"`
	Timeout        int  `json:"timeout"`
	ExpectedStatus *int `json:"expectedStatus"`
}

type TCPProbeConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type ProbeConfig struct {
	Type string           `json:"type"`
	HTTP *HTTPProbeConfig `json:"http"`
	TCP  *TCPProbeConfig  `json:"tcp"`
}

func NewHTTPProbeConfig(config *HTTPProbeConfig) *ProbeConfig {
	return &ProbeConfig{
		Type: ProbeTypeHTTP,
		HTTP: config,
	}
}

func NewTCPProbeConfig(config *TCPProbeConfig) *ProbeConfig {
	return &ProbeConfig{
		Type: ProbeTypeTCP,
		TCP:  config,
	}
}

type ProbeRequest struct {
	Config *ProbeConfig `json:"config"`
}
