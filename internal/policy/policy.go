package policy

// Policy evaluation for action capabilities

type Capability string

const (
	CapabilityDocker          Capability = "docker"
	CapabilitySystem          Capability = "system"
	CapabilityNetwork         Capability = "network"
	CapabilityIdentity        Capability = "identity"
	CapabilityCommandExecution Capability = "command-execution"
)

type Policy struct {
	AllowedCapabilities []Capability
}

func NewPolicy(capabilities []Capability) *Policy {
	return &Policy{
		AllowedCapabilities: capabilities,
	}
}

func (p *Policy) HasCapability(cap Capability) bool {
	for _, allowed := range p.AllowedCapabilities {
		if allowed == cap {
			return true
		}
	}
	return false
}

func (p *Policy) CanExecute(verb string, requiredCaps []Capability) bool {
	for _, reqCap := range requiredCaps {
		if !p.HasCapability(reqCap) {
			return false
		}
	}
	return true
}

