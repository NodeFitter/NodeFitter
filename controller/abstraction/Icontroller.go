package abstraction

import "github.com/NodeFitter/NodeFitter/comms"

type Icontroller interface {
	UpdateMemThreshold(args *comms.UpdateMemThresholdArgs, reply *comms.UpdateMemThresholdReply) error
	UpdateCPUThreshold(args *comms.UpdateCPUThresholdArgs, reply *comms.UpdateCPUThresholdReply) error
	PrintVM(args *comms.EmptyArgs, reply *comms.PrintVMReply) error
}
