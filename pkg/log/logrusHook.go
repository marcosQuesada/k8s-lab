package log

import (
	"github.com/sirupsen/logrus"
)

// globalFieldHook defines log Hook to handle global fields
type globalFieldHook struct {
	service string
	env     string
}

// NewGlobalFieldHook instantiate log Hook to handle global fields
func NewGlobalFieldHook(service string, env string) logrus.Hook {
	return &globalFieldHook{
		service: service,
		env:     env,
	}
}

// Levels defines logging level when hooks has to be fired
func (h *globalFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire all the hooks for the passed level.
func (h *globalFieldHook) Fire(entry *logrus.Entry) error {
	entry.Data["service"] = h.service
	entry.Data["env"] = h.env
	return nil
}
