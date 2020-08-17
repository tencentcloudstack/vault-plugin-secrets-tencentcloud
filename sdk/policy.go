package sdk

type Policy struct {
	// Version only allow 2.0
	Version   string      `json:"version"`
	Statement []Statement `json:"statement,omitempty"`
}

type Statement struct {
	Principal interface{} `json:"principal,omitempty"`
	Effect    string      `json:"effect,omitempty"`
	Action    interface{} `json:"action,omitempty"`
	Resource  interface{} `json:"resource,omitempty"`
	Condition interface{} `json:"condition,omitempty"`
}
