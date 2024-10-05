package common

const (
	DataTypePayload    uint8 = 0
	DataTypeConnect    uint8 = 1
	DataTypeDisconnect uint8 = 2
)

type DataHeader struct {
	Type            uint8  `json:"type"`
	ConnectionID    int64  `json:"connectionId"`
	DestinationIP   string `json:"destinationIp"`   // optional for backend to source
	DestinationPort int    `json:"destinationPort"` // optional for backend to source
}

type Data struct {
	Header  DataHeader `json:"header"`
	Payload string     `json:"payload"`
}

func NewProxyData(destinationConnId string, data []byte) {

}
