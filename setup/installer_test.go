package setup_test

import (
	"context"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/dwellersclub/contigus/log"
	"github.com/dwellersclub/contigus/models"
	"github.com/dwellersclub/contigus/setup"
)

func TestInstaller(t *testing.T) {

	pl := setup.NewWriterProgressListener(os.Stdout)
	installer := setup.NewInstaller(pl, logrus.NewEntry(log.GetLogger()))

	step := installer.Run(context.Background(), models.SetupConfig{
		Runtime: models.RuntimeEnum.Mac,
	})

	assert.Equal(t, setup.StepsEnum.Completed, step)
}
