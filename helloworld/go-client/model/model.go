package model

type Response struct {
	Args     map[string]string `json:"args"`
	Form     map[string]string `json:"form"`
	Headers  map[string]string `json:"headers"`
	Method   string            `json:"method"`
	Origin   string            `json:"origin"`
	Url      string            `json:"url"`
	Envs     map[string]string `json:"envs"`
	HostName string            `json:"host_name"`
	Body     string            `json:"body"`
}

type ResponseAny struct {
	Code int `json:"code"`
	Data any `json:"data"`
}
