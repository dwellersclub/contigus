package models

//NewHookOption Create new hook option
func NewHookOption(username, token, secret string, maxByte int64) HookOption {
	return HookOption{
		Username: username,
		Token:    token,
		Secret:   secret,
		MaxByte:  maxByte,
	}
}

//HookOption Option to configure  hook
type HookOption struct {
	Username string
	Token    string
	Secret   string
	MaxByte  int64
}
