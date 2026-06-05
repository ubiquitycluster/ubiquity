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
