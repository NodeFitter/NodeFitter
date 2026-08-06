package abstraction

import (
	"github.com/NodeFitter/NodeFitter/context"
	"github.com/NodeFitter/NodeFitter/scheduler"
)

type Ischeduler interface {
	UpdateMemoryThreshold(float64) error
	UpdateCPUThreshold(float32) error
	UpdateKubernetesCASHA(string) error
	Start(context.SchedulerConfig) error // Read context, initialize scheduler and start scheduling process
	StartScheduleProcess() error         // Start the scheduling process. Return an error if Start has not been invoked
	StopScheduleProcess() error          // Stop the scheduling process, if active
	GetVMs() []scheduler.Node
}
