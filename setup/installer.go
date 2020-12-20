package setup

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/dwellersclub/contigus/models"
)

//Step Installer step
type Step string

//Steps Installer Steps
type Steps struct {
	Start          Step
	SecurityConfig Step
	DBConfig       Step
	Continue       Step
	Completed      Step
}

//StepsEnum Steps enum
var StepsEnum = Steps{
	Start:          Step("start"),
	SecurityConfig: Step("security"),
	DBConfig:       Step("db"),
	Continue:       Step("continue"),
	Completed:      Step("completed"),
}

//Installer interface
type Installer interface {
	Run(context.Context, models.SetupConfig) Step
}

//NewInstaller Create a new Installer
func NewInstaller(progressListener ProgressListener, logger *logrus.Entry) Installer {
	return &defaultInstaller{
		progressListener: progressListener,
		logger:           logger,
		Step:             StepsEnum.Start,
		tasks: map[Step]InstallerTask{
			StepsEnum.Start: &initialSecurityTask{},
		},
	}
}

type defaultInstaller struct {
	tasks            map[Step]InstallerTask
	progressListener ProgressListener
	logger           *logrus.Entry
	Step             Step
}

func (di *defaultInstaller) Run(ctx context.Context, config models.SetupConfig) Step {
	di.logger.Info("Start install")
	for _, task := range di.tasks {
		step := task.Run(ctx, di.progressListener, config)
		if step != StepsEnum.Continue {
			return step
		}
	}
	di.logger.Info("End install")
	return StepsEnum.Completed
}
