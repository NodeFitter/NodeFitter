package abstraction

import "github.com/NodeFitter/NodeFitter/comms"

type Icontroller interface {
	GetMemThreshold(args *comms.EmptyArgs, reply *comms.GetMemThresholdReply) error
	GetCPUThreshold(args *comms.EmptyArgs, reply *comms.GetCPUThresholdReply) error
	UpdateMemThreshold(args *comms.UpdateMemThresholdArgs, reply *comms.UpdateMemThresholdReply) error
	UpdateCPUThreshold(args *comms.UpdateCPUThresholdArgs, reply *comms.UpdateCPUThresholdReply) error
	PrintVM(args *comms.EmptyArgs, reply *comms.PrintVMReply) error
}
