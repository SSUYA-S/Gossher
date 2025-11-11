package inventory

import (
	"fmt"
	"time"
)

// Ensure Host implements the interfaces
var (
	_ Entity       = (*Host)(nil)
	_ TaggedEntity = (*Host)(nil)
	_ VarContainer = (*Host)(nil)
)

// Host represents a remote server accessible via SSH.
type Host struct {
	Type DocumentType `yaml:"type"`
	// Basic identification
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`

	// SSH connection information
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`

	// Authentication - use either CredentialID (recommended) or inline auth
	CredentialID string `yaml:"credential_id,omitempty"`

	// Inline authentication (optional, overrides credential if both are set)
	User     string `yaml:"user,omitempty"`
	KeyPath  string `yaml:"key_path,omitempty"`
	Password string `yaml:"password,omitempty"`

	// Classification and metadata
	Tags []string          `yaml:"tags,omitempty"`
	Vars map[string]string `yaml:"vars,omitempty"`

	// Runtime state (not saved to YAML)
	Status       HostStatus `yaml:"-"`
	LastPingTime time.Time  `yaml:"-"`
}

// HostStatus represents the current state of a host.
type HostStatus int

const (
	HostStatusUnknown HostStatus = iota
	HostStatusOnline
	HostStatusOffline
	HostStatusConnecting
)

func (s HostStatus) String() string {
	switch s {
	case HostStatusOnline:
		return "Online"
	case HostStatusOffline:
		return "Offline"
	case HostStatusConnecting:
		return "Connecting"
	default:
		return "Unknown"
	}
}

// NewHost creates a new Host with default values.
func NewHost(id, name, address string) *Host {
	return &Host{
		Type:    TypeHost,
		ID:      id,
		Name:    name,
		Address: address,
		Port:    22,
		Tags:    []string{},
		Vars:    make(map[string]string),
		Status:  HostStatusUnknown,
	}
}

// NewHostWithCredential creates a new Host using a credential reference.
func NewHostWithCredential(id, name, address, credentialID string) *Host {
	h := NewHost(id, name, address)
	h.CredentialID = credentialID
	return h
}

// GetID Identifiable interface implementation
func (h *Host) GetID() string {
	return h.ID
}

// GetName Nameable interface implementation
func (h *Host) GetName() string {
	return h.Name
}

func (h *Host) SetName(name string) {
	h.Name = name
}

// Describable interface implementation
func (h *Host) GetDescription() string {
	return h.Description
}

func (h *Host) SetDescription(desc string) {
	h.Description = desc
}

// Validate checks if the Host has all required fields.
func (h *Host) Validate() error {
	if h.ID == "" {
		return fmt.Errorf("host ID cannot be empty")
	}
	if h.Name == "" {
		return fmt.Errorf("host %s: name cannot be empty", h.ID)
	}
	if h.Address == "" {
		return fmt.Errorf("host %s: address cannot be empty", h.ID)
	}
	if h.Port <= 0 || h.Port > 65535 {
		return fmt.Errorf("host %s: invalid port %d", h.ID, h.Port)
	}

	hasCredential := h.CredentialID != ""
	hasInlineAuth := h.User != ""

	if !hasCredential && !hasInlineAuth {
		return fmt.Errorf("host %s: must have either credential_id or user", h.ID)
	}

	return nil
}

// Clone creates a deep copy of the Host.
func (h *Host) Clone() interface{} {
	clone := *h
	clone.Tags = make([]string, len(h.Tags))
	copy(clone.Tags, h.Tags)
	clone.Vars = make(map[string]string, len(h.Vars))
	for k, v := range h.Vars {
		clone.Vars[k] = v
	}
	return &clone
}

// VarContainer interface implementation
func (h *Host) GetVar(key string) (string, bool) {
	val, ok := h.Vars[key]
	return val, ok
}

func (h *Host) SetVar(key, value string) {
	if h.Vars == nil {
		h.Vars = make(map[string]string)
	}
	h.Vars[key] = value
}

func (h *Host) GetAllVars() map[string]string {
	return h.Vars
}

// TaggedEntity interface implementation
func (h *Host) HasTag(tag string) bool {
	for _, t := range h.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (h *Host) AddTag(tag string) {
	if !h.HasTag(tag) {
		h.Tags = append(h.Tags, tag)
	}
}

func (h *Host) RemoveTag(tag string) {
	for i, t := range h.Tags {
		if t == tag {
			h.Tags = append(h.Tags[:i], h.Tags[i+1:]...)
			return
		}
	}
}

func (h *Host) GetTags() []string {
	return h.Tags
}

// SSHAddress returns the address for SSH connection in "address:port" format.
func (h *Host) SSHAddress() string {
	return fmt.Sprintf("%s:%d", h.Address, h.Port)
}

// UsesCredential returns true if this host uses a credential reference.
func (h *Host) UsesCredential() bool {
	return h.CredentialID != ""
}
