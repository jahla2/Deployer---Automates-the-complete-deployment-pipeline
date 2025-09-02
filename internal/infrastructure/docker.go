package infrastructure

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"deployer/internal/domain"
)

type DockerService struct {
	logger domain.Logger
	dryRun bool
}

func NewDockerService(logger domain.Logger, dryRun bool) *DockerService {
	return &DockerService{
		logger: logger,
		dryRun: dryRun,
	}
}

func (d *DockerService) BuildImage(imageName, version, buildPath string) error {
	if buildPath == "" {
		d.logger.Warning("No build path specified, skipping build step")
		return nil
	}

	buildDir := buildPath
	if !filepath.IsAbs(buildDir) {
		wd, _ := os.Getwd()
		buildDir = filepath.Join(wd, buildDir)
	}

	imageTag := fmt.Sprintf("%s:%s", imageName, version)
	cmd := exec.Command("docker", "build", "-t", imageTag, ".")
	cmd.Dir = buildDir

	d.logger.Info("Building in: %s", buildDir)
	d.logger.Info("Command: %s", strings.Join(cmd.Args, " "))

	if d.dryRun {
		return nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		d.logger.Error("Build output: %s", output)
		return fmt.Errorf("docker build failed: %w", err)
	}

	d.logger.Info("Image built: %s", imageTag)
	return nil
}

func (d *DockerService) TagImage(localImage, registryImage string) error {
	cmd := exec.Command("docker", "tag", localImage, registryImage)

	d.logger.Info("Command: %s", strings.Join(cmd.Args, " "))

	if d.dryRun {
		return nil
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		d.logger.Error("Tag output: %s", output)
		return fmt.Errorf("docker tag failed: %w", err)
	}

	d.logger.Info("Tagged: %s -> %s", localImage, registryImage)
	return nil
}

func (d *DockerService) LoginRegistry(host, username, password string) error {
	cmd := exec.Command("docker", "login", host, "-u", username, "-p", password)

	d.logger.Info("Command: docker login %s -u %s -p [HIDDEN]", host, username)

	if d.dryRun {
		return nil
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		d.logger.Error("Login output: %s", output)
		return fmt.Errorf("docker login failed: %w", err)
	}

	d.logger.Info("Logged into registry: %s", host)
	return nil
}

func (d *DockerService) PushImage(registryImage string) error {
	cmd := exec.Command("docker", "push", registryImage)

	d.logger.Info("Command: %s", strings.Join(cmd.Args, " "))

	if d.dryRun {
		return nil
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker push failed: %w", err)
	}

	d.logger.Info("Pushed: %s", registryImage)
	return nil
}