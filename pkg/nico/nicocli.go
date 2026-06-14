package nico

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

type CommandRunnerFunc func(ctx context.Context, name string, args ...string) ([]byte, error)

func (f CommandRunnerFunc) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return f(ctx, name, args...)
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

type NicoCLI struct {
	cfg    Config
	runner CommandRunner
}

type CLIOption func(*NicoCLI)

func WithCommandRunner(r CommandRunner) CLIOption {
	return func(c *NicoCLI) {
		if r != nil {
			c.runner = r
		}
	}
}

func WithCLIPath(path string) CLIOption {
	return func(c *NicoCLI) {
		if path != "" {
			c.cfg.CLIPath = path
		}
	}
}

func NewNicoCLI(cfg Config, opts ...CLIOption) *NicoCLI {
	cli := &NicoCLI{cfg: cfg.WithDefaults(), runner: execRunner{}}
	for _, opt := range opts {
		opt(cli)
	}
	cli.cfg = cli.cfg.WithDefaults()
	return cli
}

func (c *NicoCLI) String() string {
	return fmt.Sprintf("NicoCLI{path:%q config:%s}", c.cfg.CLIPath, c.cfg.String())
}

func (c *NicoCLI) Run(ctx context.Context, args ...string) error {
	_, err := c.runner.Run(ctx, c.cfg.CLIPath, c.args(args...)...)
	return err
}

func (c *NicoCLI) JSON(ctx context.Context, out any, args ...string) error {
	stdout, err := c.runner.Run(ctx, c.cfg.CLIPath, c.args(append(args, "--output", "json")...)...)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(stdout, out); err != nil {
		return fmt.Errorf("decode nicocli JSON: %w", err)
	}
	return nil
}

func (c *NicoCLI) args(args ...string) []string {
	var out []string
	if strings.TrimSpace(c.cfg.Org) != "" {
		out = append(out, "--org", c.cfg.Org)
	}
	if strings.TrimSpace(c.cfg.APIName) != "" {
		out = append(out, "--api", c.cfg.APIName)
	}
	out = append(out, args...)
	return out
}

func (c *NicoCLI) ListSites(ctx context.Context) ([]Site, error) {
	var out []Site
	return out, c.JSON(ctx, &out, "site", "list")
}

func (c *NicoCLI) ListMachines(ctx context.Context) ([]Machine, error) {
	var out []Machine
	return out, c.JSON(ctx, &out, "machine", "list")
}

func (c *NicoCLI) GetMachine(ctx context.Context, name string) (Machine, error) {
	var out Machine
	return out, c.JSON(ctx, &out, "machine", "get", name)
}

func (c *NicoCLI) CreateOperatingSystem(ctx context.Context, os OperatingSystem) (OperatingSystem, error) {
	var out OperatingSystem
	payload, err := jsonPayload(os)
	if err != nil {
		return out, err
	}
	return out, c.JSON(ctx, &out, "operating-system", "create", "--payload", payload)
}

func (c *NicoCLI) ListInstances(ctx context.Context) ([]Instance, error) {
	var out []Instance
	return out, c.JSON(ctx, &out, "instance", "list")
}

func (c *NicoCLI) CreateInstance(ctx context.Context, inst Instance) (Instance, error) {
	var out Instance
	payload, err := jsonPayload(inst)
	if err != nil {
		return out, err
	}
	return out, c.JSON(ctx, &out, "instance", "create", "--payload", payload)
}

func (c *NicoCLI) DeleteInstance(ctx context.Context, id string) error {
	return c.Run(ctx, "instance", "delete", id, "--yes")
}

func (c *NicoCLI) PowerMachine(ctx context.Context, machineID, state, reason string) (Task, error) {
	args := []string{"machine", "power", machineID, state}
	if strings.TrimSpace(reason) != "" {
		args = append(args, "--reason", reason)
	}
	var out Task
	return out, c.JSON(ctx, &out, args...)
}

func (c *NicoCLI) ListTasks(ctx context.Context) ([]Task, error) {
	var out []Task
	return out, c.JSON(ctx, &out, "task", "list")
}

func (c *NicoCLI) GetTask(ctx context.Context, id string) (Task, error) {
	var out Task
	return out, c.JSON(ctx, &out, "task", "get", id)
}

func (c *NicoCLI) ListMachineGPUStats(ctx context.Context) ([]MachineGPUStats, error) {
	var out []MachineGPUStats
	return out, c.JSON(ctx, &out, "machine-gpu-stats", "list")
}

func jsonPayload(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
