package generic

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/dwellersclub/contigus/models"
	"github.com/sirupsen/logrus"
)

//Read read request payload
func Read(r *http.Request, config models.HookOption) (body []byte, err error) {
	logrus.Infof("Parse generic Request")
	body, err = ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: config.MaxByte})
	return
}
