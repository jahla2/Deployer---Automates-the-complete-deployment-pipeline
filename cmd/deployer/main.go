package main

import (
    "flag"
    "fmt"
    "os"

    "deployer/internal/config"
    "deployer/internal/domain"
    "deployer/internal/infrastructure"
    "deployer/internal/ui"
    "deployer/internal/usecase"
    "deployer/pkg/logger"
)

func main() {
    var (
        configFile   = flag.String("config", "deployment.config.json", "Configuration file path")
        service      = flag.String("service", "", "Service name to deploy")
        version      = flag.String("version", "", "Version tag for the image")
        buildPath    = flag.String("build-path", "", "Path where to run docker build (optional, overrides config)")
        dryRun       = flag.Bool("dry-run", false, "Show commands without executing")
        listServices = flag.Bool("list", false, "List available services from config")
    )
    flag.Parse()

    // Initialize dependencies
    log := logger.New("deployer")
    configRepo := config.NewRepository()

    if *listServices {
        cli := ui.NewCLI(configRepo, nil, log)
        cli.ListServices(*configFile)
        return
    }

    // Interactive mode if no arguments provided
    if *service == "" || *version == "" {
        if len(os.Args) == 1 {
            // Initialize all services for interactive mode
            dockerService := infrastructure.NewDockerService(log, false)
            sshService := infrastructure.NewSSHService(domain.SSHConfig{}, log, false)
            deploymentService := usecase.NewDeploymentService(dockerService, sshService, log)
            cli := ui.NewCLI(configRepo, deploymentService, log)
            
            cli.RunInteractiveMode(*configFile)
            return
        }
        fmt.Println("Usage: deployer -service <service-name> -version <version> [-config deployment.config.json] [-build-path /path/to/build] [-dry-run]")
        fmt.Println("       deployer -list [-config deployment.config.json]")
        os.Exit(1)
    }

    // Command line mode
    config, err := configRepo.LoadConfig(*configFile)
    if err != nil {
        log.Error("Failed to load config: %v", err)
        os.Exit(1)
    }

    serviceNames := configRepo.GetServiceNames(config)
    serviceExists := false
    for _, name := range serviceNames {
        if name == *service {
            serviceExists = true
            break
        }
    }

    if !serviceExists {
        log.Error("Service '%s' not found in config. Available services: %v", *service, serviceNames)
        os.Exit(1)
    }

    // Initialize services
    dockerService := infrastructure.NewDockerService(log, *dryRun)
    sshService := infrastructure.NewSSHService(config.SSH, log, *dryRun)
    deploymentService := usecase.NewDeploymentService(dockerService, sshService, log)

    request := domain.DeploymentRequest{
        ServiceName:       *service,
        Version:          *version,
        BuildPathOverride: *buildPath,
        DryRun:           *dryRun,
    }

    if err := deploymentService.Deploy(request, config); err != nil {
        log.Error("Deployment failed: %v", err)
        os.Exit(1)
    }

    log.Info("Deployment completed successfully!")
}