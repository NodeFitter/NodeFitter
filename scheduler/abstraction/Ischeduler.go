package abstraction

import (
	"github.com/NodeFitter/NodeFitter/context"
	"github.com/NodeFitter/NodeFitter/scheduler"
)

type Ischeduler interface {
	UpdateMemoryThreshold(float64) error
	UpdateCPUThreshold(float32) error
	UpdateKubernetesCASHA(string) error
	Start(context.SchedulerConfig) error
	GetVMs() []scheduler.Node
}
