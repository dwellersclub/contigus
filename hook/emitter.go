package hook

import "github.com/dwellersclub/contigus/models"

//Emitter dispatch event received
type Emitter interface {
	Emit(*models.Event) error
	GetConfig(*models.Event) models.EmitterConfig
}
