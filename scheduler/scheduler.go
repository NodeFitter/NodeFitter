package scheduler

import (
	sysContext "context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/NodeFitter/NodeFitter/context"
	"github.com/NodeFitter/NodeFitter/utility"
	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"go.rtnl.ai/x/randstr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	tokenAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
)

var (
	ErrorNoInitializedConfig    = errors.New("error while communicating with OpenNebula: connection has not been initialized. Check configuration or Start method")
	ErrorConfigNotValid         = errors.New("error while reading the configuration file: a not valid data has been read")
	ErrorCACertPemDecoding      = errors.New("error while decoding PEM of kubernetes certificate")
	ErrorSchedulerAlreadyActive = errors.New("error while starting the scheduler process: the scheduler is already active")
	ErrorSchedulerNotActive     = errors.New("error while stopping the scheduler process: the scheduler is not active")
)

type Node struct {
	Id           int
	AvailableMem float64
	AvailableCPU float32
	VMGroupName  string
	VMTemplateId int
}

type Scheduler struct {
	// OpenNebula controller
	onController *goca.Controller

	// Kubernetes client
	k8Client *kubernetes.Clientset

	// Scheduler parameters
	freeMemoryThreshold utility.CType[float64]
	freeCPUThreshold    utility.CType[float32]
	hasBeenStarted      bool

	// Kubernetes token generation data
	kubernetesCASHA string
	k8Endpoint      string

	// Scheduler internal list of nodes (VMs)
	vms map[int]*Node

	// Scheduler timer
	ticker   *time.Ticker
	interval int
}

func (s *Scheduler) UpdateMemoryThreshold(newThresholdInMb float64) error {
	s.freeMemoryThreshold.Set(newThresholdInMb)
	return nil
}

func (s *Scheduler) UpdateCPUThreshold(newThresholdInPercentage float32) error {
	s.freeCPUThreshold.Set(newThresholdInPercentage)
	return nil
}

func (s *Scheduler) UpdateKubernetesCASHA(CApath string) error {

	if CApath == "" {

		// If path was not provided, check default locations
		CApath = "./kubernetesCert/ca.crt"

		// Check file presence
		_, err := os.Stat(CApath)

		if os.IsNotExist(err) {
			CApath = "/etc/kubernetes/pki/ca.crt"
		}
	}

	caSHA, err := s.calculateCACertSHA(CApath)

	if err != nil {
		return err
	}

	s.kubernetesCASHA = caSHA

	return nil
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

	// Create kubernetes connection
	config, err := clientcmd.BuildConfigFromFlags(
		"",
		ctx.KubernetesConfigPath,
	)

	if err != nil {
		return err
	}

	k8Client, err := kubernetes.NewForConfig(config)

	if err != nil {
		return err
	}

	s.k8Client = k8Client

	// Calculate SHA256 of certificate
	caSHA, err := s.calculateCACertSHA(ctx.KubernetesCACertificatePath)

	if err != nil {
		return err
	}

	s.kubernetesCASHA = caSHA

	// Memorize kubernetes endpoint (useful for token generation)
	s.k8Endpoint = ctx.KubernetesEndpoint

	// Create initial map of nodes
	s.vms = make(map[int]*Node)

	// Set check interval
	if ctx.SchedulerProcessInterval <= 0 {
		return ErrorConfigNotValid
	}

	s.interval = ctx.SchedulerProcessInterval

	s.hasBeenStarted = true

	s.StartScheduleProcess()

	return nil
}

func (s *Scheduler) createKubernetesJoinToken() (string, error) {

	// Generate randomly token ID and secret
	tokenID := randstr.Generate(6, tokenAlphabet)
	tokenSecret := randstr.Generate(16, tokenAlphabet)

	token := tokenID + "." + tokenSecret

	// Set expiration date of token to 24h
	expirationTime := time.Now().
		Add(24 * time.Hour).
		UTC().
		Format(time.RFC3339)

	// Create the secret to later upload to the control plane
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bootstrap-token-" + tokenID,
			Namespace: "kube-system",
			Labels: map[string]string{
				"auth-kubernetes-io/token-bootstrap": "true",
			},
		},

		Type: corev1.SecretType("bootstrap.kubernetes.io/token"),

		StringData: map[string]string{
			"token-id":                       tokenID,
			"token-secret":                   tokenSecret,
			"expiration":                     expirationTime,
			"usage-bootstrap-authentication": "true",
			"usage-bootstrap-signing":        "true",
			"auth-extra-groups":              "system:bootstrappers:kubeadm:default-node-token",
		},
	}

	// Copy token to control plane
	_, err := s.k8Client.CoreV1().
		Secrets("kube-system").
		Create(
			sysContext.Background(),
			secret,
			metav1.CreateOptions{},
		)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Scheduler) calculateCACertSHA(certPath string) (string, error) {

	// Read certificate file
	data, err := os.ReadFile(certPath)

	if err != nil {
		return "", err
	}

	// Get PEM blocks
	pemBlocks, _ := pem.Decode(data)

	if pemBlocks == nil {
		return "", ErrorCACertPemDecoding
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(pemBlocks.Bytes)

	if err != nil {
		return "", err
	}

	// Marshal key
	key, err := x509.MarshalPKIXPublicKey(cert.PublicKey)

	if err != nil {
		return "", err
	}

	// Calculate sha256
	hash := sha256.Sum256(key)

	return fmt.Sprintf("%x", hash[:]), nil
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
			s.vms[ivm.ID] = &Node{Id: ivm.ID}
		}

		// Set vm list for mem
		s.vms[ivm.ID].AvailableMem = freeMem

		// Set vm list for cpu
		s.vms[ivm.ID].AvailableCPU = float32(freeCPU)

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
		s.vms[ivm.ID].VMGroupName = vmGroupName

		// Retrieve vm template id
		templateIdPair, err := vmTemplate.GetPair("TEMPLATE_ID")

		if err != nil {
			continue
		}

		templateId, err := strconv.Atoi(templateIdPair.Value)

		if err != nil {
			continue
		}

		s.vms[ivm.ID].VMTemplateId = templateId
	}

	return nil

}

func (s *Scheduler) StartScheduleProcess() error {

	if !s.hasBeenStarted {
		return ErrorNoInitializedConfig
	}

	if s.ticker != nil {
		return ErrorSchedulerAlreadyActive
	}

	go s.startScheduleProcess()

	return nil
}

func (s *Scheduler) startScheduleProcess() {

	s.ticker = time.NewTicker(time.Duration(s.interval) * time.Second)

	for range s.ticker.C {
		s.checkAndSchedule()
	}

}

func (s *Scheduler) StopScheduleProcess() error {

	if s.ticker != nil {
		s.ticker.Stop()
	} else {
		return ErrorSchedulerNotActive
	}

	s.ticker = nil

	return nil
}

func (s *Scheduler) checkAndSchedule() error {

	// Update vm map
	s.updateVmMap()

	for _, vm := range s.vms {

		if vm.AvailableMem <= s.freeMemoryThreshold.Get() || vm.AvailableCPU <= s.freeCPUThreshold.Get() {

			templateId := vm.VMTemplateId

			tc := s.onController.Template(templateId)

			// Generate a token to make the new vm join the kubernetes cluster
			token, err := s.createKubernetesJoinToken()

			if err != nil {
				continue
			}

			joinCmd := "kubeadm join " + s.k8Endpoint + " --token " + token + " --discovery-token-ca-cert-hash sha256:" + s.kubernetesCASHA

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
			K8_JOIN_COMMAND = "` + joinCmd + `"
			]
			`

			// Instantiate the vm with the new context
			tc.Instantiate(uuid.New().String(), false, updatedContext, false)
		}
	}

	// Check if some vm can be removed
	err := s.checkAndUnschedule()

	return err
}

func (s *Scheduler) checkAndUnschedule() error {

	// Update vm map
	s.updateVmMap()

	return nil
}

func (s *Scheduler) GetVMs() []Node {
	vmList := make([]Node, len(s.vms))
	index := 0

	for _, vm := range s.vms {
		vmList[index] = *vm
		index += 1
	}

	return vmList
}
