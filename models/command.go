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

type CmdCreateAccount struct {
	Account   string      `json:"account"`
	Password   string      `json:"password"`
}

type CmdForgotPassword struct {
	Account   string      `json:"account"`
}

type CmdResetPassword struct {
	Account   string      `json:"account"`
    Password   string      `json:"password"`
    Token      string      `json:"token"`
}

type RefreshToken struct {
	RefreshToken   string      `json:"refresh_token"`
}