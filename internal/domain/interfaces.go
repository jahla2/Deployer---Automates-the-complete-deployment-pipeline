package domain

type ConfigRepository interface {
	LoadConfig(configFile string) (*Config, error)
	GetServiceNames(config *Config) []string
}

type DockerService interface {
	BuildImage(imageName, version, buildPath string) error
	TagImage(localImage, registryImage string) error
	LoginRegistry(host, username, password string) error
	PushImage(registryImage string) error
}

type SSHService interface {
	Connect(config SSHConfig) error
	RunCommand(command string) error
	RunCommandWithOutput(command string) (string, error)
}

type DeploymentService interface {
	Deploy(request DeploymentRequest, config *Config) error
}

type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Warning(msg string, args ...interface{})
	Success(msg string, args ...interface{})
}