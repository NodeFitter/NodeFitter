package comms

import "github.com/NodeFitter/NodeFitter/scheduler"

/*
This file contains the types used for communication with the CLI
*/

// Get service status
type StatusArgs struct{}
type StatusReply struct{}

// Update the Memory threshold
type UpdateMemThresholdArgs struct {
	NewThreshold float64
}
type UpdateMemThresholdReply struct {
	Success bool
}

// Update the CPU threshold
type UpdateCPUThresholdArgs struct {
	NewThreshold float32
}
type UpdateCPUThresholdReply struct {
	Success bool
}

// Print the VMs
// type PrintVMArgs struct{}
type PrintVMReply struct {
	VMs []scheduler.Node
}

// Get current CPU threshold
type GetCPUThresholdReply struct {
	Error     bool
	Threshold float32
}

// Get current Memory threshold
type GetMemThresholdReply struct {
	Error     bool
	Threshold float64
}

type EmptyArgs struct{}
type EmptyReply struct{}
