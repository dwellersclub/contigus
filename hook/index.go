package hook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/dwellersclub/contigus/models"
)

//Types Supported Hooks types
var Types = struct {
	Github string
	GitLab string
	Slack  string
	Values func() []string
}{"github", "gitlab", "slack", func() []string { return []string{"github", "gitlab", "slack"} }}

//Errors Errors raised by package
var Errors = struct {
	InvalidHook       error
	InvalidPayload    error
	InvalidEncryption error
	Values            func() []string
}{
	fmt.Errorf("invalid_hook"),
	fmt.Errorf("invalid_payload"),
	fmt.Errorf("invalid_encrypt"),
	func() []string { return []string{"invalid_hook", "invalid_payload", "invalid_encrypt"} },
}

// NewService Create new Service to use hooks
func NewService(metrics Metrics, encryptor Encryptor, serverID string) Service {
	return &defaultService{
		metrics:   metrics,
		encryptor: encryptor,
		serverID:  serverID,
	}
}

//Service Hook facade
type Service interface {
	GetHook(ID string) *hook
	MatchURL(URL string) *hook
	NewRequest(r *http.Request, hook *hook) (*models.Event, error)
	EmitRequest(*models.Event) error
}

type defaultService struct {
	metrics   Metrics
	encryptor Encryptor
	serverID  string
	emitter   Emitter
}

func (ds *defaultService) EmitRequest(rq *models.Event) error {
	return ds.emitter.Emit(rq)
}

func (ds *defaultService) MatchURL(URL string) *hook {
	return nil
}

func (ds *defaultService) GetHook(ID string) *hook {
	return nil
}

func (ds *defaultService) NewRequest(r *http.Request, hook *hook) (event *models.Event, err error) {
	t := time.Now()
	defer func() {
		elapsed := time.Since(t)
		hookType := ""
		if hook != nil {
			hookType = hook.Type
		}
		errorCode := ""
		if err != nil {
			errorCode = err.Error()
		}
		ds.metrics.IncHandled(hookType, errorCode, elapsed)
	}()

	if hook == nil {
		err = Errors.InvalidHook
		return
	}

	if !hook.Active {
		err = Errors.InvalidHook
		return
	}

	if !hook.Valid(r) {
		err = Errors.InvalidPayload
		return
	}

	// read body compress and encrypt
	encKeyID, content, encryptErr := ds.encryptor.Encrypt(r.Body)
	if encryptErr != nil {
		err = Errors.InvalidEncryption
		return
	}

	//get headers
	jsonHeader, marshallErr := json.Marshal(r.Header)
	if marshallErr != nil {
		err = Errors.InvalidPayload
		return
	}

	header, encryptErr := ds.encryptor.EncryptWithID(bytes.NewBuffer(jsonHeader), encKeyID)
	if encryptErr != nil {
		err = Errors.InvalidEncryption
		return
	}

	id := uuid.NewV4()

	return &models.Event{
		Origin: models.Origin{
			CorrolationID: id.String(),
			ServerID:      ds.serverID,
		},
		Source: models.Source{
			Date: time.Now().UnixNano(),
			ID:   hook.ID,
			Type: "web_hook",
		},
		Data: models.Data{
			Header:   header,
			Content:  content,
			EncKeyID: encKeyID,
			Type:     "json",
		},
	}, nil
}

type hook struct {
	ID     string
	Active bool
	Type   string
}

func (h *hook) Valid(r *http.Request) bool {
	// get hook validate
	return true
}
