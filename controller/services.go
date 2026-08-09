package controller

/*
This file contains the controller functions called via rpc
*/

func (c *Controller) UpdateMemThreshold(args *UpdateMemThresholdArgs, reply UpdateMemThresholdReply) error {
	// TODO: Make this return something to use in the reply
	c.scheduler.UpdateMemoryThreshold(args.NewThreshold)

	reply.Success = true

	return nil
}

func (c *Controller) UpdateCPUThreshold(args *UpdateCPUThresholdArgs, reply UpdateCPUThresholdReply) error {
	// TODO: Make this return something to use in the reply
	c.scheduler.UpdateCPUThreshold(args.NewThreshold)

	reply.Success = true

	return nil
}

func (c *Controller) PrintVM(args *EmptyArgs, reply *EmptyReply) error {
	// TODO: Make this return something to use in the reply
	return nil
}
