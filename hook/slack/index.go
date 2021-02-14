package slack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dwellersclub/contigus/models"
	"github.com/sirupsen/logrus"
)

// Errors returned by various methods.
const (
	ErrMissingHeaders   = models.SimpleError("missing headers")
	ErrExpiredTimestamp = models.SimpleError("timestamp is too old")
)

//Read read slack request payload
func Read(r *http.Request, config models.HookOption) ([]byte, error) {
	logrus.Infof("Parse slack Request")
	verifier, err := NewSecretsVerifier(r.Header, config.Secret)
	if err != nil {
		return nil, err
	}

	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))

	var body []byte
	if body, err = ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: config.MaxByte}); err != nil {
		return nil, err
	}

	if err = verifier.Ensure(); err != nil {
		return nil, err
	}

	return body, nil
}

// Signature headers
const (
	hSignature = "X-Slack-Signature"
	hTimestamp = "X-Slack-Request-Timestamp"
)

// secretsVerifier contains the information needed to verify that the request comes from Slack
type secretsVerifier struct {
	signature []byte
	hmac      hash.Hash
}

func unsafeSignatureVerifier(header http.Header, secret string) (_ secretsVerifier, err error) {
	var (
		bsignature []byte
	)

	signature := header.Get(hSignature)
	stimestamp := header.Get(hTimestamp)

	if signature == "" || stimestamp == "" {
		return secretsVerifier{}, ErrMissingHeaders
	}

	if bsignature, err = hex.DecodeString(strings.TrimPrefix(signature, "v0=")); err != nil {
		return secretsVerifier{}, err
	}

	hash := hmac.New(sha256.New, []byte(secret))
	if _, err = hash.Write([]byte(fmt.Sprintf("v0:%s:", stimestamp))); err != nil {
		return secretsVerifier{}, err
	}

	return secretsVerifier{
		signature: bsignature,
		hmac:      hash,
	}, nil
}

// NewSecretsVerifier returns a SecretsVerifier object in exchange for an http.Header object and signing secret
func NewSecretsVerifier(header http.Header, secret string) (sv secretsVerifier, err error) {
	var (
		timestamp int64
	)

	stimestamp := header.Get(hTimestamp)

	if sv, err = unsafeSignatureVerifier(header, secret); err != nil {
		return secretsVerifier{}, err
	}

	if timestamp, err = strconv.ParseInt(stimestamp, 10, 64); err != nil {
		return secretsVerifier{}, err
	}

	diff := absDuration(time.Since(time.Unix(timestamp, 0)))
	if diff > 5*time.Minute {
		return secretsVerifier{}, ErrExpiredTimestamp
	}

	return sv, err
}

func (v *secretsVerifier) Write(body []byte) (n int, err error) {
	return v.hmac.Write(body)
}

// Ensure compares the signature sent from Slack with the actual computed hash to judge validity
func (v secretsVerifier) Ensure() error {
	computed := v.hmac.Sum(nil)
	// use hmac.Equal prevent leaking timing information.
	if hmac.Equal(computed, v.signature) {
		return nil
	}
	return fmt.Errorf("Computed unexpected signature of: %s", hex.EncodeToString(computed))
}

func abs64(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

func absDuration(n time.Duration) time.Duration {
	return time.Duration(abs64(int64(n)))
}
