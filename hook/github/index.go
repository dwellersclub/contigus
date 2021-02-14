package github

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/dwellersclub/contigus/models"
	ghc "github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

const signatureHeader = "X-Hub-Signature"
const payloadFormParam = "payload"

//Read read github request payload
func Read(r *http.Request, config models.HookOption) (payload []byte, err error) {
	logrus.Infof("Parse github Request")
	return validatePayload(r, []byte(config.Secret), config.MaxByte)
}

func validatePayload(r *http.Request, secretToken []byte, maxBytes int64) (payload []byte, err error) {
	body := &bytes.Buffer{} // Raw body that GitHub uses to calculate the signature.

	switch ct := r.Header.Get("Content-Type"); ct {
	case "application/json":
		var err error
		var payload []byte
		if payload, err = ioutil.ReadAll(r.Body); err != nil {
			return nil, err
		}

		body = bytes.NewBuffer(payload)

	case "application/x-www-form-urlencoded":
		// payloadFormParam is the name of the form parameter that the JSON payload
		// will be in if a webhook has its content type set to application/x-www-form-urlencoded.

		var err error
		var payload []byte
		if payload, err = ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: maxBytes}); err != nil {
			return nil, err
		}

		// If the content type is application/x-www-form-urlencoded,
		// the JSON payload will be under the "payload" form param.
		form, err := url.ParseQuery(string(payload))
		if err != nil {
			return nil, err
		}

		body = bytes.NewBufferString(form.Get(payloadFormParam))

	default:
		return nil, fmt.Errorf("Webhook request has unsupported Content-Type %q", ct)
	}

	// Only validate the signature if a secret token exists. This is intended for
	// local development only and all webhooks should ideally set up a secret token.
	if len(secretToken) > 0 {
		sig := r.Header.Get(signatureHeader)
		if err := ghc.ValidateSignature(sig, body.Bytes(), secretToken); err != nil {
			return nil, err
		}
	}

	return payload, nil
}
