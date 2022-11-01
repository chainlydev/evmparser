package models

type Logs struct {
	Address    string      `bson:"address,omitempty"`
	Topics     []string    `bson:"topics,omitempty"`
	Data       string      `bson:"data,omitempty"`
	Index      uint64      `bson:"index,omitempty"`
	Function   string      `bson:"function,omitempty"`
	Action     string      `bson:"action,omitempty"`
	ActionType string      `bson:"action_type,omitempty"`
	Token      string      `bson:"token,omitempty"`
	Parameters interface{} `bson:"parameters,omitempty"`
	Removed    bool        `bson:"removed,omitempty"`
}
