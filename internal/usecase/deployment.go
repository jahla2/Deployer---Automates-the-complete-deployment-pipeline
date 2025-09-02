package usecase

import (
	"fmt"
	"strings"

	"deployer/internal/domain"
)

type DeploymentService struct {
	dockerService domain.DockerService
	sshService    domain.SSHService
	logger        domain.Logger
}

func NewDeploymentService(dockerService domain.DockerService, sshService domain.SSHService, logger domain.Logger) *DeploymentService {
	return &DeploymentService{
		dockerService: dockerService,
		sshService:    sshService,
		logger:        logger,
	}
}

func (d *DeploymentService) Deploy(request domain.DeploymentRequest, config *domain.Config) error {
	serviceConfig, exists := config.Services[request.ServiceName]
	if !exists {
		return fmt.Errorf("service '%s' not found in config", request.ServiceName)
	}

	if request.BuildPathOverride != "" {
		serviceConfig.BuildPath = request.BuildPathOverride
	}

	steps := []struct {
		name string
		fn   func() error
	}{
		{"Building Docker image", func() error { return d.buildImage(serviceConfig, request.Version) }},
		{"Tagging image for registry", func() error { return d.tagImage(serviceConfig, request.Version, config.Registry) }},
		{"Logging into registry", func() error { return d.loginRegistry(config.Registry) }},
		{"Pushing image to registry", func() error { return d.pushImage(serviceConfig, request.Version, config.Registry) }},
		{"Connecting to remote server", func() error { return d.sshService.Connect(config.SSH) }},
		{"Pulling image on remote", func() error { return d.pullImageRemote(serviceConfig, request.Version, config) }},
		{"Stopping existing container", func() error { return d.stopContainer(serviceConfig.ContainerName) }},
		{"Removing existing container", func() error { return d.removeContainer(serviceConfig.ContainerName) }},
		{"Running new container", func() error { return d.runContainer(serviceConfig, request.Version, config.Registry) }},
		{"Verifying container mounts", func() error { return d.checkContainerStatus(serviceConfig.ContainerName, serviceConfig.HealthTimeout) }},
	}

	for i, step := range steps {
		progress := int(float64(i) / float64(len(steps)) * 100)
		d.showProgressBar(progress)
		d.logger.Info("[%d/%d] %s", i+1, len(steps), step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("step '%s' failed: %w", step.name, err)
		}
		progress = int(float64(i+1) / float64(len(steps)) * 100)
		d.showProgressBar(progress)
		d.logger.Success("[%d/%d] COMPLETED: %s", i+1, len(steps), step.name)
	}

	return nil
}

func (d *DeploymentService) showProgressBar(progress int) {
	const (
		width = 50
		green = "\033[32m"
		blue  = "\033[34m"
		reset = "\033[0m"
		bold  = "\033[1m"
	)
	
	filled := int(float64(width) * float64(progress) / 100.0)
	filledBar := strings.Repeat("█", filled)
	emptyBar := strings.Repeat("░", width-filled)
	
	var color string
	if progress == 100 {
		color = green
	} else {
		color = blue
	}
	
	fmt.Printf("\r%s%s[%s%s%s%s] %s%d%%%s%s", 
		bold, color, green, filledBar, reset, emptyBar, bold, progress, reset, reset)
	if progress == 100 {
		fmt.Println()
	}
}

func (d *DeploymentService) buildImage(serviceConfig domain.DeployConfig, version string) error {
	return d.dockerService.BuildImage(serviceConfig.ImageName, version, serviceConfig.BuildPath)
}

func (d *DeploymentService) tagImage(serviceConfig domain.DeployConfig, version string, registry domain.RegistryConfig) error {
	localImage := fmt.Sprintf("%s:%s", serviceConfig.ImageName, version)
	registryImage := fmt.Sprintf("%s/%s:%s", registry.Host, serviceConfig.ImageName, version)
	return d.dockerService.TagImage(localImage, registryImage)
}

func (d *DeploymentService) loginRegistry(registry domain.RegistryConfig) error {
	return d.dockerService.LoginRegistry(registry.Host, registry.Username, registry.Password)
}

func (d *DeploymentService) pushImage(serviceConfig domain.DeployConfig, version string, registry domain.RegistryConfig) error {
	registryImage := fmt.Sprintf("%s/%s:%s", registry.Host, serviceConfig.ImageName, version)
	return d.dockerService.PushImage(registryImage)
}

func (d *DeploymentService) pullImageRemote(serviceConfig domain.DeployConfig, version string, config *domain.Config) error {
	registryImage := fmt.Sprintf("%s/%s:%s", config.Registry.Host, serviceConfig.ImageName, version)

	commands := []string{
		fmt.Sprintf("docker login %s -u %s -p %s", config.Registry.Host, config.Registry.Username, config.Registry.Password),
		fmt.Sprintf("docker pull %s", registryImage),
	}

	for _, cmd := range commands {
		if err := d.sshService.RunCommand(cmd); err != nil {
			return fmt.Errorf("remote command failed '%s': %w", cmd, err)
		}
	}

	d.logger.Info("Image pulled on remote: %s", registryImage)
	return nil
}

func (d *DeploymentService) stopContainer(containerName string) error {
	cmd := fmt.Sprintf("docker stop %s || true", containerName)
	if err := d.sshService.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	d.logger.Info("Container stopped: %s", containerName)
	return nil
}

func (d *DeploymentService) removeContainer(containerName string) error {
	cmd := fmt.Sprintf("docker rm %s || true", containerName)
	if err := d.sshService.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	d.logger.Info("Container removed: %s", containerName)
	return nil
}

func (d *DeploymentService) runContainer(serviceConfig domain.DeployConfig, version string, registry domain.RegistryConfig) error {
	registryImage := fmt.Sprintf("%s/%s:%s", registry.Host, serviceConfig.ImageName, version)

	cmd := fmt.Sprintf("docker run -d --name %s %s %s",
		serviceConfig.ContainerName,
		serviceConfig.DockerRunArgs,
		registryImage)

	if err := d.sshService.RunCommand(cmd); err != nil {
		return fmt.Errorf("failed to run container: %w", err)
	}

	d.logger.Info("Container started: %s", serviceConfig.ContainerName)
	return nil
}

func (d *DeploymentService) checkContainerStatus(containerName string, healthTimeout int) error {
	cmd := "docker ps --format 'table {{.Names}}\\t{{.Status}}\\t{{.Ports}}'"

	output, err := d.sshService.RunCommandWithOutput(cmd)
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}

	d.logger.Info("Container status:\n%s", output)

	// Check container mounts
	mountCmd := fmt.Sprintf("docker inspect %s --format='{{range .Mounts}}{{println .Source \"->\" .Destination}}{{end}}'", containerName)
	mountOutput, err := d.sshService.RunCommandWithOutput(mountCmd)
	if err != nil {
		return fmt.Errorf("failed to check container mounts: %w", err)
	}

	d.logger.Info("Container mounts:\n%s", mountOutput)

	return nil
}