package abstraction

import (
	"github.com/NodeFitter/NodeFitter/context"
)

type Ischeduler interface {
	UpdateMemoryThreshold(float64)
	UpdateCPUThreshold(float32)
	Start(context.SchedulerConfig) error
	PrintVm()
}
