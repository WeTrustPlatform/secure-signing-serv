package sss

// TxPayload to unmarshal requests JSON encoded body
type TxPayload struct {
	To       string `json:"to"`       // destination address, optional
	Value    string `json:"value"`    // amount to be transferred, optional
	GasPrice string `json:"gasPrice"` // price of the gas unit, mandatory
	Data     string `json:"data"`     // hex encoded data, optional
}

// RetryPayload to unmarshal payload when retrying a transaction
type RetryPayload struct {
	Hash     string `json:"hash"`     // transaction hash
	GasPrice string `json:"gasPrice"` // new gas price
}
