package hook

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	github "github.com/dwellersclub/contigus/hook/github"
	slack "github.com/dwellersclub/contigus/hook/slack"
	"github.com/dwellersclub/contigus/models"
	uuid "github.com/satori/go.uuid"
)

//Types Supported Hooks types
var Types = struct {
	Github  string
	GitLab  string
	Slack   string
	Generic string
	Values  func() []string
}{"github", "gitlab", "slack", "generic", func() []string { return []string{"github", "gitlab", "slack", "generic"} }}

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
func NewService(metrics Metrics, encryptor Encryptor, serverID string, eventPathIndexer Indexer, repository Repository) Service {
	return &defaultService{
		metrics:          metrics,
		encryptor:        encryptor,
		serverID:         serverID,
		eventPathIndexer: eventPathIndexer,
		repository:       repository,
	}
}

//Service Hook facade
type Service interface {
	GetHook(ID string) *Hook
	MatchURL(URL string) *Hook
	Read(r *http.Request, hook *Hook) (*models.Event, error)
	Emit(*models.Event) error
}

type defaultService struct {
	metrics          Metrics
	encryptor        Encryptor
	serverID         string
	emitter          Emitter
	eventPathIndexer Indexer
	repository       Repository
}

func (ds *defaultService) Emit(rq *models.Event) error {
	return ds.emitter.Emit(rq)
}

func (ds *defaultService) MatchURL(URL string) *Hook {
	return nil
}

func (ds *defaultService) GetHook(ID string) *Hook {
	hooks := ds.repository.GetHooks()

	for _, hook := range hooks {
		if string(hook.ID) == ID {
			hk := NewHookFromConfig(hook)
			return &hk
		}
	}

	return nil
}

func (ds *defaultService) Read(r *http.Request, hook *Hook) (event *models.Event, err error) {

	t := time.Now()
	defer func() {
		elapsed := time.Since(t)
		hookType := ""
		if hook != nil {
			hookType = hook.hookType
		}
		errorCode := ""
		if err != nil {
			errorCode = err.Error()
		}
		ds.metrics.IncHandled(hookType, string(hook.ID()), errorCode, elapsed)
	}()

	if hook == nil {
		err = Errors.InvalidHook
		return
	}

	if !hook.active {
		err = Errors.InvalidHook
		return
	}

	var data *bytes.Buffer
	ct := r.Header.Get("Content-Type")
	if ds.eventPathIndexer != nil && hook.IsIndexFields() && ct == "application/json" {
		data, err = ds.readWithDataIndex(r, hook)
	} else {
		data, err = ds.readPayload(r, hook)
	}

	// read body compress and encrypt
	encKeyID, content, encryptErr := ds.encryptor.Encrypt(data)
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
			ID:   string(hook.id),
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

func (ds *defaultService) readWithDataIndex(r *http.Request, hook *Hook) (*bytes.Buffer, error) {
	var data *bytes.Buffer
	var err error
	pr, pw := io.Pipe()

	// we need to wait for everything to be done
	wg := sync.WaitGroup{}
	wg.Add(2)

	body := ioutil.NopCloser(io.TeeReader(&io.LimitedReader{R: r.Body, N: 10240}, pw))
	r.Body = pr

	go func() {

		defer wg.Done()
		defer pw.Close()

		bufSize := 50
		bufIndex := 0
		keyBuffer := make([]byte, bufSize)
		currentKeyPath := make([]string, 5)
		previousLevel := 0
		level := 0

		for {

			br := bufio.NewReaderSize(body, 128)
			buf := make([]byte, 128)

			_, err := br.Read(buf)
			if err == io.EOF {
				break
			}

			for _, b := range buf {

				keyBuffer[bufIndex] = b

				if b == ':' {

					//find previous token should be " or space , if anything else skip
					// TODO handle escape
					i := findPreviousToken('"', keyBuffer, bufIndex)
					if i > -1 {
						j := findPreviousToken('"', keyBuffer, i)
						if j > -1 {
							content := string(keyBuffer[j+1 : i])
							if content != ": " {
								currentKeyPath[level] = content
								if previousLevel != level {
									previousLevel = level
								}
								ds.eventPathIndexer.Index(string(hook.ID()), fmt.Sprintf("%s.%s", strings.Join(currentKeyPath[0:level], "."), currentKeyPath[level]))
							}
						}
					}
					bufIndex = -1
				} else if b == '{' {
					level++
					bufIndex = -1
				} else if b == '}' {
					level--
					bufIndex = -1
				}

				//put in buffer
				bufIndex++
				if bufIndex == bufSize {
					bufIndex = 0
				}

			}

		}
	}()

	if data, err = hook.Read(r); err != nil {
		wg.Done()
		return nil, err
	}

	wg.Done()
	wg.Wait()

	return data, nil
}

func findPreviousToken(b byte, data []byte, index int) int {
	found := -1
	for i := index - 1; i >= 0; i-- {
		if data[i] == '"' {
			return i
		} else if data[i] != ' ' {
			found++
		}
	}
	return -1
}

func (ds *defaultService) readPayload(r *http.Request, hook *Hook) (*bytes.Buffer, error) {
	var data *bytes.Buffer
	var err error

	if data, err = hook.Read(r); err != nil {
		return nil, err
	}

	return data, nil
}

//NewHookFromConfig Create a new Hook from config
func NewHookFromConfig(config models.HookConfig) Hook {
	var reader hookRead

	switch config.Type {
	case "github":
		reader = github.Read
	case "slack":
		reader = slack.Read
	}

	options := models.HookOption{}

	return Hook{
		id:          config.ID,
		name:        config.Name,
		hookType:    config.Type,
		active:      config.Active,
		indexFields: config.IndexFields,
		reader:      reader,
		options:     options,
	}
}

//NewHook Create a new Hook
func NewHook(id models.UUID, hookType string, active bool) Hook {
	return NewHookFromConfig(models.HookConfig{
		ID:     id,
		Active: active,
		Type:   hookType,
	})
}

//Hook Web hook
type Hook struct {
	id          models.UUID
	name        string
	active      bool
	hookType    string
	reader      hookRead
	options     models.HookOption
	indexFields bool
}

//Read read and validate request payload
func (h *Hook) Read(r *http.Request) (*bytes.Buffer, error) {
	data, err := h.reader(r, h.options)
	return bytes.NewBuffer(data), err
}

//IsActive is Hook enabled
func (h *Hook) IsActive() bool {
	return h.active
}

//Name get hook name
func (h *Hook) Name() string {
	return h.name
}

//IsIndexFields is Indexing fields enabled
func (h *Hook) IsIndexFields() bool {
	return h.indexFields
}

//ID Hook id
func (h *Hook) ID() models.UUID {
	return h.id
}

//Type Hook type
func (h *Hook) Type() string {
	return h.hookType
}

type hookRead func(r *http.Request, config models.HookOption) ([]byte, error)
