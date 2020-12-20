package github

import (
	"bytes"
	"io/ioutil"
	"net/http"

	ghc "github.com/google/go-github/v32/github"
)

//NewHookOption Create new github hook option
func NewHookOption(username, token, secret string) HookOption {
	return HookOption{
		username: username,
		token:    token,
		secret:   secret,
	}
}

//HookOption Option to configure  hook
type HookOption struct {
	username string
	token    string
	secret   string
}

type hook struct {
	config HookOption
}

func (gh *hook) Valid(r *http.Request) bool {
	body, err := ghc.ValidatePayload(r, []byte(gh.config.secret))
	if err != nil {
		return false
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return true
}
