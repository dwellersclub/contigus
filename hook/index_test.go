package hook_test

import (
	"bytes"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dwellersclub/contigus/hook"
	"github.com/dwellersclub/contigus/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewGithubRequest(t *testing.T) {
	hookService := hook.NewService(&TestMetrics{}, hook.NewEncryptor(), "myServer", &TestIndexer{})

	data, err := ioutil.ReadFile("fixtures/index/github.txt")
	if err != nil {
		t.Fail()
		return
	}

	r := httptest.NewRequest("POST", "https://be312ae689a1.ngrok.io", bytes.NewBuffer(data))

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("User-Agent", "GitHub-Hookshot/cd078e3")
	r.Header.Add("X-GitHub-Delivery", "10ee6750-4774-11eb-844c-1ea1abc45c5e")
	r.Header.Add("X-GitHub-Event", "ping")
	r.Header.Add("X-GitHub-Hook-ID", "271236003")
	r.Header.Add("X-GitHub-Hook-Installation-Target-ID", "150732310")
	r.Header.Add("X-GitHub-Hook-Installation-Target-Type", "repository")
	r.Header.Add("X-Hub-Signature", "sha1=320b727f4f94b56ea78956b1323a1e9b808a1b5b")
	r.Header.Add("X-Hub-Signature-256", "sha256=a39823e4902df78d6e82c8d8989ec10462d50d64a9ca70e0b53eacc225b43a6d")

	config := models.HookConfig{
		ID:          models.UUID("my_id"),
		Type:        "github",
		Active:      true,
		IndexFields: true,
		Metas:       map[string]string{"secret": "qwertyuiop"},
	}
	hook := hook.NewHookFromConfig(config)
	_, err = hookService.Read(r, &hook)

	assert.NoError(t, err)
}

func TestNewSlackRequest(t *testing.T) {
	hookService := hook.NewService(&TestMetrics{}, hook.NewEncryptor(), "myServer", &TestIndexer{})
	r := httptest.NewRequest("POST", "/home", bytes.NewBuffer([]byte{}))
	hook := hook.NewHook("my_id", "slack", true)
	hookService.Read(r, &hook)
}

type TestMetrics struct{}

func (tm *TestMetrics) IncHandled(hookType string, ID string, errorCode string, elapseTime time.Duration) {
}

type TestIndexer struct{}

func (ti *TestIndexer) Index(eventType string, fieldPath string) {
	logrus.Infof("eventType %s  fieldPath %s", eventType, fieldPath)
}
