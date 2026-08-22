package controller

import (
	"log"

	"github.com/NodeFitter/NodeFitter/comms"
)

/*
This file contains the controller functions called via rpc
*/

func (c *Controller) GetMemThreshold(args *comms.EmptyArgs, reply *comms.GetMemThresholdReply) error {
	log.Println("[CLI] Received a VM memory threshold getter request")

	value, err := c.scheduler.GetCurrentMemoryThreshold()

	if err != nil {
		log.Printf("[CLI] Memory getter request failed: %s\n", err)
		reply.Error = true
		reply.Threshold = 0
		return err
	}

	reply.Error = false
	reply.Threshold = value

	return nil
}

func (c *Controller) GetCPUThreshold(args *comms.EmptyArgs, reply *comms.GetCPUThresholdReply) error {
	log.Println("[CLI] Received a VM CPU threshold getter request")

	value, err := c.scheduler.GetCurrentCPUThreshold()

	if err != nil {
		log.Printf("[CLI] CPU getter request failed: %s\n", err)
		reply.Error = true
		reply.Threshold = 0
		return err
	}

	reply.Error = false
	reply.Threshold = value

	return nil
}

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

// Return the list of all active VMs in the cluster
func (c *Controller) PrintVM(args *comms.EmptyArgs, reply *comms.PrintVMReply) error {
	log.Println("[CLI] Received a VM status request")
	reply.VMs = c.scheduler.GetVMs()
	return nil
}

func (c *Controller) Start(args *comms.EmptyArgs, reply *comms.EmptyReply) error {
	log.Println("[CLI] Received a start request")
	err := c.scheduler.StartScheduleProcess()

	if err != nil {
		log.Printf("[ERROR] Failed to start the service: %s\n", err)
		return err
	}

	log.Println("[CLI] Successfully started the scheduling process")
	return nil
}

func (c *Controller) Stop(args *comms.EmptyArgs, reply *comms.EmptyReply) error {
	log.Println("[CLI] Received a stop request")
	err := c.scheduler.StopScheduleProcess()

	if err != nil {
		log.Printf("[ERROR] Failed to stop the service: %s\n", err)
		return err
	}

	log.Println("[CLI] Successfully stopped the scheduling process")
	return nil
}
