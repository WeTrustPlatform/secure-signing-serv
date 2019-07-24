package sss

// TxPayload to unmarshal requests JSON encoded body
type TxPayload struct {
	To       string `json:"to"`       // destination address, optional
	Value    string `json:"value"`    // amount to be transferred, optional
	GasPrice string `json:"gasPrice"` // price of the gas unit, mandatory
	Data     string `json:"data"`     // hex encoded data, optional
}

// RetryPayload to unmarshal payload when patching a transaction
type RetryPayload []PatchOperation

// PatchOperation is an operation of an HTTP PATCH
type PatchOperation struct {
	Op    string `json:"op"`    // operation
	Path  string `json:"path"`  // path to operate on
	Value string `json:"value"` // new value
}
