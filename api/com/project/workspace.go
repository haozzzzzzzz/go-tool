/**
workspace contain different env config
*/
package project

import (
	"github.com/haozzzzzzzz/go-rapid-development/utils/yaml"
	"github.com/sirupsen/logrus"
	"os"
)

const ProjectFileMode os.FileMode = os.ModePerm ^ 0111 // rw-rw-rw

const ProjectDirMode os.FileMode = os.ModePerm ^ 0111 // rw-rw-rw

type Workspace struct {
	ProjectDir    string                    `yaml:"project_dir" validate:"required"`
	WorkspaceName string                    `yaml:"workspace_name" validate:"required"`
	Stage         Stage                     `yaml:"stage" validate:"required"`
	Platform      string                    `yaml:"platform" validate:"required"`
	GoRoot        string                    `yaml:"go_root" validate:"required"`
	GoPath        string                    `yaml:"go_path" validate:"required"`
	GoOS          string                    `yaml:"go_os" validate:"required"`
	GOArch        string                    `yaml:"go_arch" validate:"required"`
	Service       map[string]*ServiceParams `yaml:"service" validate:"required"`
}

func (m *Workspace) GetService(serviceType string) *ServiceParams {
	return m.Service[serviceType]
}

type ServiceParams struct {
	ServiceEcsParams
	ServiceDocParams

	ServiceName     string `yaml:"service_name" validate:"required"`
	Description     string `yaml:"description"`
	HealthCheckPath string `yaml:"health_check_path"`
	ServicePort     int64  `yaml:"service_port"` // 服务端口
}

type ServiceEcsParams struct {
	EcsServicePortBlueGreen int64 `yaml:"ecs_service_port_blue_green"` // ecs蓝绿部署端口
	EcsTaskMemory           int64 `yaml:"ecs_task_memory"`             // ecs 内存M
}

type ServiceDocParams struct {
	DocServiceName string `yaml:"doc_service_name"`
	DocDescription string `yaml:"doc_description"`
}

func LoadWorkspace(workspacePath string) (workspace *Workspace, err error) {
	workspace = &Workspace{}
	err = yaml.ReadYamlFromFile(workspacePath, workspace)
	if nil != err {
		logrus.Errorf("read workspace from file failed. error: %s.", err)
		return
	}
	return
}
