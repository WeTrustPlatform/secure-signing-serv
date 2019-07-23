package sss

// Payload to unmarshal requests JSON encoded body
type Payload struct {
	To       string `json:"to"`       // destination address, optional
	Value    string `json:"value"`    // amount to be transferred, optional
	GasPrice string `json:"gasPrice"` // price of the gas unit, mandatory
	Data     string `json:"data"`     // hex encoded data, optional
}
