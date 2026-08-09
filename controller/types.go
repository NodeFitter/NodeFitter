package controller

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
type PrintVMArgs struct{}
type PrintVMReply struct{}

type EmptyArgs struct{}
type EmptyReply struct{}
