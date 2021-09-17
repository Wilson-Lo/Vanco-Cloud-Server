package models

type Command struct {
	Etag   string      `json:"etag"`
	Time   string      `json:"time"`
	Body   interface{} `json:"body"`
	Sign   string      `json:"sign"`
	To     string      `json:"to"`
	Extra  string      `json:"extra"`
	Method string      `json:"method"`
}
