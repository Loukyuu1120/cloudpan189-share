package filetasklog

import (
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/utils"
)

type Tracker struct {
	id    int64
	start time.Time
}

func (t *Tracker) GetID() int64 {
	return t.id
}

func (t *Tracker) Cost() time.Duration {
	return time.Since(t.start)
}

func (t *Tracker) WithCost() utils.Field {
	return utils.WithField("duration", time.Since(t.start).Milliseconds())
}

func newTracker(id int64, start time.Time) *Tracker {
	return &Tracker{id: id, start: start}
}
