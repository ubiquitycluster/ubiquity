package nico

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestNicoCLIAdapterInvokesBinaryAndParsesJSON(t *testing.T) {
	var gotName string
	var gotArgs []string
	runner := CommandRunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return []byte(`[{"name":"site-a"}]`), nil
	})
	cli := NewNicoCLI(Config{Org: "acme", APIName: "nico", TokenCommand: "pass show nico"}.WithDefaults(), WithCommandRunner(runner), WithCLIPath("/usr/bin/nicocli"))
	var out []Site
	if err := cli.JSON(context.Background(), &out, "site", "list"); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	if gotName != "/usr/bin/nicocli" {
		t.Fatalf("binary = %q", gotName)
	}
	wantArgs := []string{"--org", "acme", "--api", "nico", "site", "list", "--output", "json"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
	if len(out) != 1 || out[0].Name != "site-a" {
		t.Fatalf("decoded output = %#v", out)
	}
}

func TestNicoCLIAdapterRedactsAndPropagatesErrors(t *testing.T) {
	boom := errors.New("boom")
	cli := NewNicoCLI(Config{Token: "secret-token", TokenCommand: "pass show nico/token"}.WithDefaults(), WithCommandRunner(CommandRunnerFunc(func(context.Context, string, ...string) ([]byte, error) {
		return nil, boom
	})))
	if err := cli.Run(context.Background(), "site", "list"); !errors.Is(err, boom) {
		t.Fatalf("Run error = %v, want boom", err)
	}
	if s := cli.String(); strings.Contains(s, "secret-token") || strings.Contains(s, "pass show") {
		t.Fatalf("String leaked secret data: %s", s)
	}
}

func TestNicoCLIJSONRejectsInvalidJSON(t *testing.T) {
	cli := NewNicoCLI(Config{}.WithDefaults(), WithCommandRunner(CommandRunnerFunc(func(context.Context, string, ...string) ([]byte, error) {
		return []byte(`not-json`), nil
	})))
	var raw json.RawMessage
	if err := cli.JSON(context.Background(), &raw, "site", "list"); err == nil {
		t.Fatalf("invalid JSON should fail")
	}
}
