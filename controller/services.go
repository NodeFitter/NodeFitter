package controller

import (
	"log"

	"github.com/NodeFitter/NodeFitter/comms"
)

/*
This file contains the controller functions called via rpc
*/

func (c *Controller) UpdateMemThreshold(
	args *comms.UpdateMemThresholdArgs,
	reply *comms.UpdateMemThresholdReply,
) error {
	log.Println("[CLI] Received a memory update request")
	err := c.scheduler.UpdateMemoryThreshold(args.NewThreshold)

	if err != nil {
		log.Printf("[CLI] Memory update request failed: %s\n", err)
		reply.Success = false
		return err
	}

	log.Printf("[CLI] Successful memory update to %f\n", args.NewThreshold)
	reply.Success = true
	return nil
}

func (c *Controller) UpdateCPUThreshold(args *comms.UpdateCPUThresholdArgs, reply *comms.UpdateCPUThresholdReply) error {
	log.Println("[CLI] Received a CPU update request")
	err := c.scheduler.UpdateCPUThreshold(args.NewThreshold)

	if err != nil {
		log.Printf("[CLI] CPU update request failed: %s\n", err)
		reply.Success = false
		return err
	}

	log.Printf("[CLI] Successful CPU update to %f\n", args.NewThreshold)
	reply.Success = true
	return nil
}

func (c *Controller) PrintVM(args *comms.EmptyArgs, reply *comms.PrintVMReply) error {
	log.Println("[CLI] Received a VM status request")
	reply.VMs = c.scheduler.GetVMs()
	return nil
}
