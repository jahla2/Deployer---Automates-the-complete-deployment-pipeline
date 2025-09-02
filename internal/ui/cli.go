package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"deployer/internal/domain"
)

type CLI struct {
	configRepo domain.ConfigRepository
	deployment domain.DeploymentService
	logger     domain.Logger
}

func NewCLI(configRepo domain.ConfigRepository, deployment domain.DeploymentService, logger domain.Logger) *CLI {
	return &CLI{
		configRepo: configRepo,
		deployment: deployment,
		logger:     logger,
	}
}

func (c *CLI) RunInteractiveMode(configFile string) {
	const (
		bold   = "\033[1m"
		cyan   = "\033[36m"
		reset  = "\033[0m"
		green  = "\033[32m"
		red    = "\033[31m"
		yellow = "\033[33m"
	)
	
	fmt.Printf("%s%sDeployer v0.1 - Repsoft Limited%s\n", bold, cyan, reset)
	fmt.Printf("%s===============================%s\n", cyan, reset)

	config, err := c.configRepo.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("%sERROR: Failed to load config: %v%s\n", red, err, reset)
		fmt.Println("Press Enter to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	scanner := bufio.NewScanner(os.Stdin)

	serviceList := c.configRepo.GetServiceNames(config)

	fmt.Println("\nAvailable Services:")
	for i, name := range serviceList {
		fmt.Printf("  [%d] %s\n", i+1, name)
	}

	fmt.Print("\nSelect service (enter number): ")
	scanner.Scan()
	selection := strings.TrimSpace(scanner.Text())

	var serviceName string
	if num := c.parseNumber(selection); num > 0 && num <= len(serviceList) {
		serviceName = serviceList[num-1]
		fmt.Printf("%sSelected: %s%s\n", green, serviceName, reset)
	} else {
		fmt.Printf("ERROR: Invalid selection '%s'! Please enter a number between 1 and %d\n", selection, len(serviceList))
		fmt.Println("Press Enter to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	fmt.Print("Enter version (e.g., 1.0.0): ")
	scanner.Scan()
	version := strings.TrimSpace(scanner.Text())

	if version == "" {
		fmt.Println("ERROR: Version cannot be empty!")
		fmt.Println("Press Enter to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		return
	}

	serviceConfig := config.Services[serviceName]
	fmt.Printf("Current build path: %s\n", serviceConfig.BuildPath)
	fmt.Print("Override build path? (Enter for default, or specify new path): ")
	scanner.Scan()
	buildPathOverride := strings.Trim(strings.TrimSpace(scanner.Text()), `"`)

	fmt.Print("Dry run mode? (y/n): ")
	scanner.Scan()
	dryRunInput := strings.ToLower(strings.TrimSpace(scanner.Text()))
	dryRun := dryRunInput == "y" || dryRunInput == "yes"

	if buildPathOverride != "" {
		fmt.Printf("%sUsing custom build path: %s%s\n", yellow, buildPathOverride, reset)
	}

	fmt.Printf("\nStarting deployment of %s:%s\n", serviceName, version)
	if dryRun {
		fmt.Println("MODE: DRY RUN - No actual changes will be made")
	}
	fmt.Printf("Build path: %s\n", serviceConfig.BuildPath)
	fmt.Println("===============================")

	request := domain.DeploymentRequest{
		ServiceName:       serviceName,
		Version:          version,
		BuildPathOverride: buildPathOverride,
		DryRun:           dryRun,
	}

	if err := c.deployment.Deploy(request, config); err != nil {
		fmt.Printf("%sERROR: Deployment failed: %v%s\n", red, err, reset)
	} else {
		fmt.Printf("%s%sSUCCESS: Deployment completed successfully!%s\n", bold, green, reset)
	}

	fmt.Println("\nPress Enter to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func (c *CLI) ListServices(configFile string) {
	config, err := c.configRepo.LoadConfig(configFile)
	if err != nil {
		c.logger.Error("Failed to load config: %v", err)
		return
	}

	fmt.Printf("Available services in %s:\n", configFile)
	for name, service := range config.Services {
		fmt.Printf("  - %s\n", name)
		fmt.Printf("    Image: %s\n", service.ImageName)
		fmt.Printf("    Container: %s\n", service.ContainerName)
		fmt.Printf("    Build Path: %s\n", service.BuildPath)
		fmt.Printf("    Docker Args: %s\n", service.DockerRunArgs)
		fmt.Println()
	}
}

func (c *CLI) parseNumber(s string) int {
	num, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return num
}