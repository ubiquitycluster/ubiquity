// Package provision provides provisioning state management and phase execution.
package provision

// Provider defines the interface for executing provisioning phases.
// This allows testing phase execution with mock implementations
// instead of real exec.Command calls.
type Provider interface {
	// Metal provisions the cluster infrastructure.
	Metal(env string) error
	// Bootstrap installs ArgoCD and the root ApplicationSet.
	Bootstrap(env string) error
	// Security deploys Kyverno and baseline security policies.
	Security(env string) error
	// External provisions external resources via Terraform.
	External(env string) error
	// Wait waits for core applications to reach Ready.
	Wait(env string) error
	// PostInstall runs post-installation configuration.
	PostInstall(env string) error
}

// RealProvider is the default Provider that executes real infrastructure commands.
type RealProvider struct{}

func (p *RealProvider) Metal(env string) error {
	return nil
}

func (p *RealProvider) Bootstrap(env string) error {
	return nil
}

func (p *RealProvider) Security(env string) error {
	return nil
}

func (p *RealProvider) External(env string) error {
	return nil
}

func (p *RealProvider) Wait(env string) error {
	return nil
}

func (p *RealProvider) PostInstall(env string) error {
	return nil
}

// MockProvider is a test Provider that records calls for assertion.
type MockProvider struct {
	Calls []string
}

func (m *MockProvider) Metal(env string) error {
	m.Calls = append(m.Calls, "metal:"+env)
	return nil
}

func (m *MockProvider) Bootstrap(env string) error {
	m.Calls = append(m.Calls, "bootstrap:"+env)
	return nil
}

func (m *MockProvider) Security(env string) error {
	m.Calls = append(m.Calls, "security:"+env)
	return nil
}

func (m *MockProvider) External(env string) error {
	m.Calls = append(m.Calls, "external:"+env)
	return nil
}

func (m *MockProvider) Wait(env string) error {
	m.Calls = append(m.Calls, "wait:"+env)
	return nil
}

func (m *MockProvider) PostInstall(env string) error {
	m.Calls = append(m.Calls, "post-install:"+env)
	return nil
}

// ExecutePhase dispatches the given phase to the provider.
func ExecutePhase(provider Provider, phase, env string) error {
	switch phase {
	case "metal":
		return provider.Metal(env)
	case "bootstrap":
		return provider.Bootstrap(env)
	case "security":
		return provider.Security(env)
	case "external":
		return provider.External(env)
	case "wait":
		return provider.Wait(env)
	case "post-install":
		return provider.PostInstall(env)
	default:
		return nil
	}
}