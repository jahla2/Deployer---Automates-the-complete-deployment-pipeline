package infrastructure

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"deployer/internal/domain"
	"golang.org/x/crypto/ssh"
)

type SSHService struct {
	config       domain.SSHConfig
	activeConfig domain.SSHConfig
	logger       domain.Logger
	dryRun       bool
}

func NewSSHService(config domain.SSHConfig, logger domain.Logger, dryRun bool) *SSHService {
	return &SSHService{
		config: config,
		logger: logger,
		dryRun: dryRun,
	}
}

func (s *SSHService) Connect(config domain.SSHConfig) error {
	s.logger.Info("Connecting to: %s@%s:%d", config.Username, config.Host, config.Port)

	if s.dryRun {
		return nil
	}

	// Store the config for use in subsequent commands
	s.activeConfig = config

	client, err := s.getSSHClientWithConfig(config)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}
	defer client.Close()

	s.logger.Info("SSH connection established")
	return nil
}

func (s *SSHService) RunCommand(command string) error {
	_, err := s.RunCommandWithOutput(command)
	return err
}

func (s *SSHService) RunCommandWithOutput(command string) (string, error) {
	s.logger.Info("Remote command: %s", strings.ReplaceAll(command, s.activeConfig.Host, "[HOST]"))

	if s.dryRun {
		return "DRY RUN: Command would be executed", nil
	}

	client, err := s.getSSHClientWithConfig(s.activeConfig)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nSTDERR: " + stderr.String()
	}

	return output, err
}

func (s *SSHService) getSSHClientWithConfig(config domain.SSHConfig) (*ssh.Client, error) {
	var auth []ssh.AuthMethod

	if config.Password != "" {
		auth = append(auth, ssh.Password(config.Password))
	}

	if config.KeyFile != "" {
		key, err := ioutil.ReadFile(config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("unable to parse private key: %w", err)
		}

		auth = append(auth, ssh.PublicKeys(signer))
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	return ssh.Dial("tcp", addr, sshConfig)
}

func (s *SSHService) getSSHClient() (*ssh.Client, error) {
	var auth []ssh.AuthMethod

	if s.config.Password != "" {
		auth = append(auth, ssh.Password(s.config.Password))
	}

	if s.config.KeyFile != "" {
		key, err := ioutil.ReadFile(s.config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("unable to parse private key: %w", err)
		}

		auth = append(auth, ssh.PublicKeys(signer))
	}

	config := &ssh.ClientConfig{
		User:            s.config.Username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	return ssh.Dial("tcp", addr, config)
}