package context

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	ErrorNoConfigFileInPath           = errors.New("error while reading configuration file for OpenNebula: missing config.yml file in given path (default in ./config/config.yml)")
	ErrorWhileReadingConfigFile       = errors.New("error while reading configuration file for OpenNebula: cannot complete file read")
	ErrorWhileUnmarshallingConfigFile = errors.New("error while reading configuration file for OpenNebula: cannot complete file unmarshalling")
)

var (
	DefaultSchedulerConfigPath  = "./config/schedulerConfig.yml"
	DefaultControllerConfigPath = "./config/controllerConfig.yml"
)

type ControllerConfig struct {
}

type SchedulerConfig struct {
	User          string `yaml:"user"`
	Password      string `yaml:"password"`
	Endpoint      string `yaml:"endpoint"`
	ResScriptPath string `yaml:"res_script_path"`

	KubernetesConfigPath        string `yaml:"kubernetes_config_path"`
	KubernetesCACertificatePath string `yaml:"kubernetes_ca_certificate_path"`
	KubernetesEndpoint          string `yaml:"kubernetes_endpoint"`

	FreeMemoryThreshold      float64 `yaml:"free_ram_threshold"`
	FreeCPUThreshold         float32 `yaml:"free_CPU_threshold"`
	SchedulerProcessInterval int     `yaml:"schedule_check_interval"`
	PreserveVMTimeout        int     `yaml:"schedule_preserve_VM_timeout"`
}

type InitialContext struct {
	ControllerContext ControllerConfig
	SchedulerContext  SchedulerConfig
}

func (ic *InitialContext) readSchedulerConfig() error {

	pathWithFileName := DefaultSchedulerConfigPath

	// Check file presence
	_, err := os.Stat(pathWithFileName)

	if os.IsNotExist(err) {
		return ErrorNoConfigFileInPath
	}

	// Read config
	fileBytes, err := os.ReadFile(pathWithFileName)

	if err != nil {
		return ErrorWhileReadingConfigFile
	}

	err = yaml.Unmarshal(fileBytes, &ic.SchedulerContext)

	if err != nil {
		return ErrorWhileUnmarshallingConfigFile
	}

	return nil
}

func (ic *InitialContext) readControllerCofig() error {
	pathWithFileName := DefaultControllerConfigPath

	// Check file presence
	_, err := os.Stat(pathWithFileName)

	if os.IsNotExist(err) {
		return ErrorNoConfigFileInPath
	}

	// Read config
	fileBytes, err := os.ReadFile(pathWithFileName)

	if err != nil {
		return ErrorWhileReadingConfigFile
	}

	err = yaml.Unmarshal(fileBytes, &ic.SchedulerContext)

	if err != nil {
		return ErrorWhileUnmarshallingConfigFile
	}

	return nil
}

func (ic *InitialContext) ReadConfig() error {
	err := ic.readSchedulerConfig()

	if err != nil {
		return err
	}

	err = ic.readControllerCofig()

	if err != nil {
		return err
	}

	return nil
}
