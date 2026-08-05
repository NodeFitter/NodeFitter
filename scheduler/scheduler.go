package scheduler

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/NodeFitter/NodeFitter/context"
	"github.com/NodeFitter/NodeFitter/utility"
	"github.com/OpenNebula/one/src/oca/go/src/goca"
)

var (
	ErrorNoInitializedConfig = errors.New("error while communicating with OpenNebula: connection has not been initialized. Check configuration or Start method")
)

type node struct {
	id           int
	availableMem float64
	availableCPU float32
	vmGroupName  string
	vmTemplateId int
}

type Scheduler struct {
	// OpenNebula controller
	onController *goca.Controller

	// Scheduler parameters
	freeMemoryThreshold utility.CType[float64]
	freeCPUThreshold    utility.CType[float32]
	hasBeenStarted      bool

	// Scheduler internal list of nodes (VMs)
	vms map[int]*node
}

func (s *Scheduler) UpdateMemoryThreshold(newThresholdInMb float64) {
	s.freeMemoryThreshold.Set(newThresholdInMb)
}

func (s *Scheduler) UpdateCPUThreshold(newThresholdInPercentage float32) {
	s.freeCPUThreshold.Set(newThresholdInPercentage)
}

func (s *Scheduler) Start(ctx context.SchedulerConfig) error {

	// Set CPU and memory threshold
	s.freeCPUThreshold.Set(ctx.FreeCPUThreshold)
	s.freeMemoryThreshold.Set(ctx.FreeMemoryThreshold)

	// Create OpenNebula connection
	client, err := goca.NewClientFromConfig(
		goca.NewConfig(ctx.User, ctx.Password, ctx.Endpoint),
	)

	if err != nil {
		return err
	}

	s.onController = goca.NewController(client)

	// Create initial map of nodes
	s.vms = make(map[int]*node)

	s.hasBeenStarted = true
	s.checkAndSchedule()

	return nil
}

func (s *Scheduler) getVMGroupById(id string) string {

	vmGroupName := ""

	vmGroupId, err := strconv.Atoi(id)
	if err == nil {
		vmGroupController := s.onController.VMGroup(vmGroupId)
		vmGroup, err := vmGroupController.Info(true)
		if err == nil {
			vmGroupName = vmGroup.Name
		}
	}

	return vmGroupName
}

func (s *Scheduler) getQtOfVMsByTemplateId(templateId int) (int, error) {

	counter := 0

	vms, err := s.onController.VMs().Info(-2)

	if err != nil {
		return 0, nil
	}

	for _, ivm := range vms.VMs {

		// Retrieve all data
		vmInfo, err := s.onController.VM(ivm.ID).Info(false)

		if err != nil {
			continue
		}

		vmTemplate := vmInfo.Template

		// Retrieve vm template id

		templateIdPair, err := vmTemplate.GetPair("TEMPLATE_ID")

		if err != nil {
			continue
		}

		vmTemplateId, err := strconv.Atoi(templateIdPair.Value)

		if err != nil {
			continue
		}

		if vmTemplateId == templateId {
			counter += 1
		}

	}

	return counter, nil
}

func (s *Scheduler) updateVmMap() error {

	if !s.hasBeenStarted {
		return ErrorNoInitializedConfig
	}

	vms, err := s.onController.VMs().Info(-2)

	if err != nil {
		return err
	}

	for _, ivm := range vms.VMs {

		// Retrieve all data
		vmInfo, err := s.onController.VM(ivm.ID).Info(false)

		if err != nil {
			continue
		}

		// Retrieve user template where variable is stored
		userTemplate := vmInfo.UserTemplate
		vmTemplate := vmInfo.Template

		// Retrieve free memory
		freeMemPair, err := userTemplate.GetPair("FREE_MEM")
		freeMem := math.MaxFloat64

		if err == nil {
			// If data was available, set to respective value
			freeMem, err = strconv.ParseFloat(freeMemPair.Value, 64)

			if err != nil {
				freeMem = math.MaxFloat64
			}
		}

		// Retrieve free memory
		freeCPUPair, err := userTemplate.GetPair("FREE_CPU")
		freeCPU := math.MaxFloat32

		if err == nil {
			// If data was available, set to respective value
			freeCPU, err = strconv.ParseFloat(freeCPUPair.Value, 32)

			if err != nil {
				freeCPU = math.MaxFloat64
			}
		}

		if s.vms[ivm.ID] == nil {
			s.vms[ivm.ID] = &node{id: ivm.ID}
		}

		// Set vm list for mem
		s.vms[ivm.ID].availableMem = freeMem

		// Set vm list for cpu
		s.vms[ivm.ID].availableCPU = float32(freeCPU)

		// Retrieve vm group name
		vmGroupName := ""
		vmGroupVector, err := vmTemplate.GetVector("VMGROUP")
		if err == nil {
			vmGroupPair, err := vmGroupVector.GetPair("VMGROUP_ID")
			if err == nil {
				vmGroupName = s.getVMGroupById(vmGroupPair.Value)
			}
		}

		// Retrieve vm group name
		s.vms[ivm.ID].vmGroupName = vmGroupName

		// Retrieve vm template id
		templateIdPair, err := vmTemplate.GetPair("TEMPLATE_ID")

		if err != nil {
			continue
		}

		templateId, err := strconv.Atoi(templateIdPair.Value)

		if err != nil {
			continue
		}

		s.vms[ivm.ID].vmTemplateId = templateId
	}

	return nil

}

func (s *Scheduler) checkAndSchedule() error {

	// Update vm map
	s.updateVmMap()

	for _, vm := range s.vms {

		if vm.availableMem <= s.freeMemoryThreshold.Get() || vm.availableCPU <= s.freeCPUThreshold.Get() {

			fmt.Println(vm.availableMem, "<=", s.freeMemoryThreshold.Get(), vm.availableCPU, "<=", s.freeCPUThreshold.Get())

			templateId := vm.vmTemplateId

			tc := s.onController.Template(templateId)

			// Recover template context: it is necessary so it is possible to inject the kubeadm join command that later the vm could execute
			templateInfo, err := tc.Info(true, true)

			if err != nil {
				continue
			}

			oldContextVector, err := templateInfo.Template.GetVector("CONTEXT")

			if err != nil {
				continue
			}

			oldContext := oldContextVector.String()

			oldContext = oldContext[:len(oldContext)-2] // Remove final ] char

			updatedContext := oldContext + `,
			K8_JOIN_COMMAND = "kubeadm join --token test --discovery-token-ca-cert-hash sha256:sha_sum"
			]
			`

			// Instantiate the vm with the new context
			tc.Instantiate("clone of "+strconv.Itoa(vm.id), false, updatedContext, false)
		}
	}

	return nil
}

func (s *Scheduler) PrintVm() {
	fmt.Println("--- VMs ---")
	for i, vm := range s.vms {
		fmt.Println(i, " - ", vm.availableCPU, vm.availableMem, vm.vmGroupName, vm.vmTemplateId)
	}
}
