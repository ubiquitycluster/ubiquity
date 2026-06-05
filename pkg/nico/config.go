package nico

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

const redactedValue = "<redacted>"

type Mode string

const (
	ModeLive    Mode = "live"
	ModeMock    Mode = "mock"
	ModeOffline Mode = "offline"
)

type Config struct {
	BaseURL              string
	Org                  string
	SiteName             string
	APIName              string
	Token                string
	TokenEnv             string
	TokenCommand         string
	ConfigPath           string
	Mode                 Mode
	CLIPath              string
	AllowUnauthenticated bool
}

type TokenResolver interface {
	ResolveToken(context.Context, []string) (string, error)
}

type TokenResolverFunc func(context.Context, []string) (string, error)

func (f TokenResolverFunc) ResolveToken(ctx context.Context, argv []string) (string, error) {
	return f(ctx, argv)
}

type execTokenResolver struct{}

func (execTokenResolver) ResolveToken(ctx context.Context, argv []string) (string, error) {
	if len(argv) == 0 {
		return "", fmt.Errorf("empty token command")
	}
	out, err := exec.CommandContext(ctx, argv[0], argv[1:]...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func (c Config) WithDefaults() Config {
	if c.APIName == "" {
		c.APIName = "nico"
	}
	if c.Mode == "" {
		c.Mode = ModeLive
	}
	if c.CLIPath == "" {
		c.CLIPath = "nicocli"
	}
	if c.TokenEnv == "" {
		c.TokenEnv = "UBIQUITY_NICO_TOKEN"
	}
	return c
}

func (c Config) Validate() error {
	c = c.WithDefaults()
	switch c.Mode {
	case ModeLive, ModeMock, ModeOffline:
	default:
		return fmt.Errorf("unsupported nico mode %q", c.Mode)
	}
	if c.Mode != ModeLive {
		return nil
	}
	if strings.TrimSpace(c.BaseURL) == "" {
		return fmt.Errorf("nico base URL is required in live mode")
	}
	u, err := url.ParseRequestURI(c.BaseURL)
	if err != nil || u.Scheme == "" || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("invalid nico base URL %q", c.BaseURL)
	}
	if strings.TrimSpace(c.Org) == "" {
		return fmt.Errorf("nico org is required in live mode")
	}
	if strings.TrimSpace(c.Token) == "" && strings.TrimSpace(c.TokenCommand) == "" && !c.AllowUnauthenticated {
		return fmt.Errorf("nico token or token command is required in live mode")
	}
	return nil
}

func (c Config) ResolveToken(ctx context.Context, resolver TokenResolver) (string, error) {
	c = c.WithDefaults()
	if strings.TrimSpace(c.Token) != "" {
		return strings.TrimSpace(c.Token), nil
	}
	if c.TokenEnv != "" {
		if token := strings.TrimSpace(os.Getenv(c.TokenEnv)); token != "" {
			return token, nil
		}
	}
	if strings.TrimSpace(c.TokenCommand) == "" {
		return "", nil
	}
	argv, err := safeTokenCommandArgs(c.TokenCommand)
	if err != nil {
		return "", err
	}
	if resolver == nil {
		resolver = execTokenResolver{}
	}
	token, err := resolver.ResolveToken(ctx, argv)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(token), nil
}

func safeTokenCommandArgs(command string) ([]string, error) {
	if strings.ContainsAny(command, ";&|`$<>(){}[]*?!\n\r") {
		return nil, fmt.Errorf("unsafe token command: shell metacharacters are not allowed")
	}
	argv := strings.Fields(command)
	if len(argv) == 0 {
		return nil, fmt.Errorf("empty token command")
	}
	return argv, nil
}

func (c Config) String() string { return c.redacted(false) }

func (c Config) Debug() string { return c.redacted(true) }

func (c Config) redacted(debug bool) string {
	c = c.WithDefaults()
	token := ""
	if c.Token != "" {
		token = redactedValue
	}
	tokenCommand := ""
	if c.TokenCommand != "" {
		tokenCommand = redactedValue
	}
	if debug {
		return fmt.Sprintf("Config{BaseURL:%q Org:%q SiteName:%q APIName:%q Mode:%q CLIPath:%q ConfigPath:%q Token:%q TokenEnv:%q TokenCommand:%q}", c.BaseURL, c.Org, c.SiteName, c.APIName, c.Mode, c.CLIPath, c.ConfigPath, token, c.TokenEnv, tokenCommand)
	}
	return fmt.Sprintf("nico config baseURL=%q org=%q site=%q api=%q mode=%q token=%q tokenEnv=%q tokenCommand=%q", c.BaseURL, c.Org, c.SiteName, c.APIName, c.Mode, token, c.TokenEnv, tokenCommand)
}
