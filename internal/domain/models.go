package domain

type DeployConfig struct {
	ServiceName   string `json:"service_name"`
	ImageName     string `json:"image_name"`
	Registry      string `json:"registry"`
	BuildPath     string `json:"build_path"`
	ContainerName string `json:"container_name"`
	DockerRunArgs string `json:"docker_run_args"`
	HealthTimeout int    `json:"health_timeout"`
}

type RegistryConfig struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SSHConfig struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	KeyFile  string `json:"key_file"`
}

type Config struct {
	Registry RegistryConfig            `json:"registry"`
	SSH      SSHConfig                 `json:"ssh"`
	Services map[string]DeployConfig   `json:"services"`
}

type DeploymentRequest struct {
	ServiceName       string
	Version          string
	BuildPathOverride string
	DryRun           bool
}