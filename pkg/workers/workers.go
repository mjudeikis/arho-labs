package workers

import (
	"github.com/mjudeikis/osa-labs/pkg/api"
)

type Workers interface {
	Get() (*api.Worker, error)
	Create() error
}
