package setup

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dwellersclub/contigus/models"
	"github.com/dwellersclub/contigus/utils"
)

//NewWriterProgressListener Create a new progress listener backed by a write
// if write is of type http.Flush a flush call will be called for every write
func NewWriterProgressListener(w io.Writer) ProgressListener {
	flusher, flushable := w.(http.Flusher)

	return &writerProgressListener{
		writer:    w,
		flushable: flushable,
		flusher:   flusher,
	}
}

//writerProgressListener Progress listener that writes to a Writer interface
type writerProgressListener struct {
	writer    io.Writer
	flushable bool
	flusher   http.Flusher
}

func (wplc *writerProgressListener) onProgress(p Progress) {
	wplc.writer.Write(p.ToJSONArray())
	if wplc.flushable {
		wplc.flusher.Flush()
	}
}

//ProgressListener Progress listener
type ProgressListener interface {
	onProgress(Progress)
}

//Progress Progress details
type Progress struct {
	TaskName string
	Percent  int
	Message  string
}

//ToJSONArray Return an non escaped json array representation
func (p *Progress) ToJSONArray() []byte {
	return []byte(fmt.Sprintf("[\"%s\", %d , \"%s\"]", p.TaskName, p.Percent, p.Message))
}

//InstallerTask interface
type InstallerTask interface {
	Run(context.Context, ProgressListener, models.SetupConfig) Step
}

type initialSecurityTask struct{}

func (ist *initialSecurityTask) Run(ctx context.Context, listener ProgressListener, config models.SetupConfig) Step {

	if !ist.validateConfig(config) {
		return StepsEnum.SecurityConfig
	}

	if config.Runtime == models.RuntimeEnum.Kubernetes {
		//key := fmt.Sprintf("secrets-%s", config.ReleaseName)

		// check if a secrets exists
	} else {
		key := fmt.Sprintf("secrets-%s-%s", config.Namespace, config.ReleaseName)
		filename := filepath.FromSlash(fmt.Sprintf("%s%s", config.RootDir, key))
		if utils.Exists(filename) {
			//load details and verify it from the file
		} else {

			encryptionKey := config.EncryptionKey

			if len(encryptionKey) == 0 {
				// auto generate key
				// encryptionKey = ?????
			}

			// create secure file
			ioutil.WriteFile(filename, []byte(encryptionKey), os.ModePerm)

		}
		// check if a file has been created with default secret
	}
	return StepsEnum.Continue
}

func (ist *initialSecurityTask) validateConfig(config models.SetupConfig) bool {
	return true
}
