package controller

import "github.com/NodeFitter/NodeFitter/comms"

/*
This file contains the controller functions called via rpc
*/

func (c *Controller) UpdateMemThreshold(
	args *comms.UpdateMemThresholdArgs,
	reply *comms.UpdateMemThresholdReply,
) error {
	err := c.scheduler.UpdateMemoryThreshold(args.NewThreshold)

	if err != nil {
		reply.Success = false
		return err
	}

	reply.Success = true
	return nil
}

func (c *Controller) UpdateCPUThreshold(args *comms.UpdateCPUThresholdArgs, reply *comms.UpdateCPUThresholdReply) error {
	err := c.scheduler.UpdateCPUThreshold(args.NewThreshold)

	if err != nil {
		reply.Success = false
		return err
	}

	reply.Success = true

	return nil
}

func (c *Controller) PrintVM(args *comms.EmptyArgs, reply *comms.PrintVMReply) error {
	reply.VMs = c.scheduler.GetVMs()
	return nil
}
