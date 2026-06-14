package nico

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

const defaultHTTPTimeout = 30 * time.Second

type Client struct {
	cfg        Config
	httpClient *http.Client
}

type ClientOption func(*Client)

func WithHTTPClient(h *http.Client) ClientOption {
	return func(c *Client) {
		if h != nil {
			c.httpClient = h
		}
	}
}

func NewClient(cfg Config, opts ...ClientOption) (*Client, error) {
	cfg = cfg.WithDefaults()
	if cfg.Token == "" {
		token, err := cfg.ResolveToken(context.Background(), nil)
		if err != nil {
			return nil, err
		}
		cfg.Token = token
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	c := &Client{cfg: cfg, httpClient: &http.Client{Timeout: defaultHTTPTimeout}}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

type Site struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Machine struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	SiteID     string `json:"siteId,omitempty"`
	SiteName   string `json:"siteName,omitempty"`
	PowerState string `json:"powerState,omitempty"`
	Status     string `json:"status,omitempty"`
	LastAction string `json:"lastAction,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type PowerRequest struct {
	State  string `json:"state"`
	Reason string `json:"reason,omitempty"`
}

type OperatingSystem struct {
	ID         string              `json:"id,omitempty"`
	Name       string              `json:"name,omitempty"`
	APIVersion string              `json:"apiVersion,omitempty"`
	Kind       string              `json:"kind,omitempty"`
	Metadata   ObjectMetadata      `json:"metadata,omitempty"`
	Spec       OperatingSystemSpec `json:"spec,omitempty"`
}

type ObjectMetadata struct {
	Name   string            `json:"name,omitempty"`
	Labels map[string]string `json:"labels,omitempty"`
}

type OperatingSystemSpec struct {
	Family       string            `json:"family,omitempty"`
	Version      string            `json:"version,omitempty"`
	Architecture string            `json:"architecture,omitempty"`
	ImageURL     string            `json:"imageURL,omitempty"`
	Checksum     string            `json:"checksum,omitempty"`
	Provenance   string            `json:"provenance,omitempty"`
	IPXEScript   string            `json:"ipxeScript,omitempty"`
	UserData     string            `json:"userData,omitempty"`
	Labels       map[string]string `json:"labels,omitempty"`
}

type Instance struct {
	ID              string            `json:"id,omitempty"`
	Name            string            `json:"name,omitempty"`
	MachineID       string            `json:"machineId,omitempty"`
	NodeName        string            `json:"nodeName,omitempty"`
	Status          string            `json:"status,omitempty"`
	OSID            string            `json:"osId,omitempty"`
	OSImage         string            `json:"osImage,omitempty"`
	InstanceTypeRef string            `json:"instanceTypeRef,omitempty"`
	GPUProfile      string            `json:"gpuProfile,omitempty"`
	JoinProfile     string            `json:"joinProfile,omitempty"`
	MachineSelector map[string]string `json:"machineSelector,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	LastAction      string            `json:"lastAction,omitempty"`
	Reason          string            `json:"reason,omitempty"`
}

type MachineGPUStats struct {
	MachineID string `json:"machineId,omitempty"`
	Count     int    `json:"count,omitempty"`
}

func (c *Client) ListSites(ctx context.Context) ([]Site, error) {
	var out []Site
	return out, c.get(ctx, c.resourcePath("site"), &out)
}

func (c *Client) ListMachines(ctx context.Context) ([]Machine, error) {
	var out []Machine
	return out, c.get(ctx, c.resourcePath("machine"), &out)
}

func (c *Client) GetMachine(ctx context.Context, name string) (Machine, error) {
	var out Machine
	return out, c.get(ctx, c.resourcePath("machine", name), &out)
}

func (c *Client) PowerMachine(ctx context.Context, machineID, state, reason string) (Task, error) {
	var out Task
	return out, c.do(ctx, http.MethodPost, c.resourcePath("machine", machineID, "power"), PowerRequest{State: state, Reason: reason}, &out)
}

func (c *Client) ListOperatingSystems(ctx context.Context) ([]OperatingSystem, error) {
	var out []OperatingSystem
	return out, c.get(ctx, c.resourcePath("operating-system"), &out)
}

func (c *Client) CreateOperatingSystem(ctx context.Context, os OperatingSystem) (OperatingSystem, error) {
	var out OperatingSystem
	return out, c.do(ctx, http.MethodPost, c.resourcePath("operating-system"), os, &out)
}

func (c *Client) ListInstances(ctx context.Context) ([]Instance, error) {
	var out []Instance
	return out, c.get(ctx, c.resourcePath("instance"), &out)
}

func (c *Client) CreateInstance(ctx context.Context, inst Instance) (Instance, error) {
	var out Instance
	return out, c.do(ctx, http.MethodPost, c.resourcePath("instance"), inst, &out)
}

func (c *Client) DeleteInstance(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, c.resourcePath("instance", id), nil, nil)
}

func (c *Client) ListTasks(ctx context.Context) ([]Task, error) {
	var out []Task
	return out, c.get(ctx, c.resourcePath("task"), &out)
}

func (c *Client) ListMachineGPUStats(ctx context.Context) ([]MachineGPUStats, error) {
	var out []MachineGPUStats
	return out, c.get(ctx, c.resourcePath("machine-gpu-stats"), &out)
}

func (c *Client) resourcePath(resource string, parts ...string) string {
	segments := []string{"v2", "org", c.cfg.Org, c.cfg.APIName, resource}
	segments = append(segments, parts...)
	escaped := make([]string, 0, len(segments))
	for _, s := range segments {
		escaped = append(escaped, url.PathEscape(s))
	}
	return "/" + strings.Join(escaped, "/")
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

func (c *Client) do(ctx context.Context, method, path string, in any, out any) error {
	if _, err := url.Parse(c.cfg.BaseURL); err != nil {
		return fmt.Errorf("parse base URL: %w", err)
	}
	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + path
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nico request failed: %s %s: %w", method, c.redact(endpoint), err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("nico request failed: %s %s: status %d: %s", method, c.redact(endpoint), resp.StatusCode, c.redact(strings.TrimSpace(string(respBody))))
	}
	if out == nil || len(strings.TrimSpace(string(respBody))) == 0 {
		return nil
	}
	if err := decodeNICOResponse(respBody, out); err != nil {
		return fmt.Errorf("decode nico response: %w", err)
	}
	return nil
}

func decodeNICOResponse(respBody []byte, out any) error {
	if err := json.Unmarshal(respBody, out); err == nil {
		return nil
	} else if !isPointerToSlice(out) {
		return err
	}
	var envelope struct {
		Items json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return err
	}
	if len(envelope.Items) == 0 {
		return fmt.Errorf("expected array or items envelope")
	}
	return json.Unmarshal(envelope.Items, out)
}

func isPointerToSlice(v any) bool {
	rv := reflect.ValueOf(v)
	return rv.IsValid() && rv.Kind() == reflect.Pointer && !rv.IsNil() && rv.Elem().Kind() == reflect.Slice
}

func (c *Client) redact(s string) string {
	for _, secret := range []string{c.cfg.Token, c.cfg.TokenCommand} {
		if strings.TrimSpace(secret) != "" {
			s = strings.ReplaceAll(s, secret, redactedValue)
		}
	}
	return s
}
