package inventory

import "fmt"

// Ensure Credential implements the interfaces
var (
	_ Entity = (*Credential)(nil)
)

// Credential represents SSH authentication information that can be shared across multiple hosts.
type Credential struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`

	User     string `yaml:"user"`
	KeyPath  string `yaml:"key_path,omitempty"`
	Password string `yaml:"password,omitempty"`

	Passphrase string `yaml:"passphrase,omitempty"`
}

// CredentialType represents the authentication method.
type CredentialType int

const (
	CredentialTypeKey CredentialType = iota
	CredentialTypePassword
)

// NewCredential creates a new Credential with basic information.
func NewCredential(id, name, user string) *Credential {
	return &Credential{
		ID:   id,
		Name: name,
		User: user,
	}
}

// Identifiable interface implementation
func (c *Credential) GetID() string {
	return c.ID
}

// Nameable interface implementation
func (c *Credential) GetName() string {
	return c.Name
}

func (c *Credential) SetName(name string) {
	c.Name = name
}

// Describable interface implementation
func (c *Credential) GetDescription() string {
	return c.Description
}

func (c *Credential) SetDescription(desc string) {
	c.Description = desc
}

// Validate checks if the Credential has valid configuration.
func (c *Credential) Validate() error {
	if c.ID == "" {
		return fmt.Errorf("credential ID cannot be empty")
	}
	if c.Name == "" {
		return fmt.Errorf("credential %s: name cannot be empty", c.ID)
	}
	if c.User == "" {
		return fmt.Errorf("credential %s: user cannot be empty", c.ID)
	}

	if c.KeyPath == "" && c.Password == "" {
		return fmt.Errorf("credential %s: must have either key_path or password", c.ID)
	}

	return nil
}

// Clone creates a deep copy of the Credential.
func (c *Credential) Clone() interface{} {
	clone := *c
	return &clone
}

// Type returns the authentication type of this credential.
func (c *Credential) Type() CredentialType {
	if c.KeyPath != "" {
		return CredentialTypeKey
	}
	return CredentialTypePassword
}
