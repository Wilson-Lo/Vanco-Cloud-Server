package models

type Command struct {
	Etag   string      `json:"etag"`
	Time   string      `json:"time"`
	Body   string      `json:"body"`
	Sign   string      `json:"sign"`
	To     string      `json:"to"`
	Extra  string      `json:"extra"`
	Method string      `json:"method"`
}

type AccountObject struct {
	Account   string      `json:"account"`
	Password   string      `json:"password"`
}

type ForgotPasswordObject struct {
	Account   string      `json:"account"`
}

type ResetPasswordObject struct {
	Account   string      `json:"account"`
    Password   string      `json:"password"`
    Token      string      `json:"token"`
}

type RefreshTokenObject struct {
	RefreshToken   string      `json:"refresh_token"`
}

type DeviceInfoObject struct {
	Mac   string      `json:"mac"`
	Name   string      `json:"name"`
	Time   string      `json:"time"`
	Type   string      `json:"type"`
}