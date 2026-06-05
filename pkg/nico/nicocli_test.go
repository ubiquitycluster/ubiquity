package nico

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func TestNicoCLIImplementsLifecycleClientWithJSONPayloads(t *testing.T) {
	var calls [][]string
	cli := NewNicoCLI(Config{Org: "acme", Token: "secret-token"}.WithDefaults(), WithCommandRunner(CommandRunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		joined := strings.Join(args, " ")
		if strings.Contains(joined, "secret-token") {
			t.Fatalf("nicocli args leaked token: %q", joined)
		}
		switch {
		case strings.Contains(joined, "machine list"):
			return []byte(`[{"id":"machine-1","name":"cn01"}]`), nil
		case strings.Contains(joined, "instance create"):
			return []byte(`{"id":"inst-1","nodeName":"cn01","machineId":"machine-1"}`), nil
		case strings.Contains(joined, "machine power"):
			return []byte(`{"id":"task-power-1","status":"running","machineId":"machine-1","action":"power off"}`), nil
		case strings.Contains(joined, "instance delete"):
			return []byte(``), nil
		default:
			return []byte(`[]`), nil
		}
	})))
	machines, err := cli.ListMachines(context.Background())
	if err != nil || len(machines) != 1 || machines[0].Name != "cn01" {
		t.Fatalf("ListMachines = %#v, %v", machines, err)
	}
	inst, err := cli.CreateInstance(context.Background(), Instance{NodeName: "cn01", MachineID: "machine-1"})
	if err != nil || inst.ID != "inst-1" {
		t.Fatalf("CreateInstance = %#v, %v", inst, err)
	}
	task, err := cli.PowerMachine(context.Background(), "machine-1", "off", "maintenance")
	if err != nil || task.ID != "task-power-1" {
		t.Fatalf("PowerMachine = %#v, %v", task, err)
	}
	if err := cli.DeleteInstance(context.Background(), "inst-1"); err != nil {
		t.Fatalf("DeleteInstance: %v", err)
	}
	joinedCalls := fmt.Sprint(calls)
	for _, want := range []string{"machine list", "instance create", "--payload", "machine power", "--reason maintenance", "instance delete inst-1 --yes"} {
		if !strings.Contains(joinedCalls, want) {
			t.Fatalf("calls %s missing %q", joinedCalls, want)
		}
	}
}
