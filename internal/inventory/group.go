package inventory

import "fmt"

// Ensure Group implements the interfaces
var (
	_ Entity       = (*Group)(nil)
	_ VarContainer = (*Group)(nil)
)

// Group represents a collection of hosts that can be managed together.
type Group struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	HostIDs     []string          `yaml:"host_ids"`
	Vars        map[string]string `yaml:"vars,omitempty"`

	ChildGroupNames []string `yaml:"child_groups,omitempty"`
}

// NewGroup creates a new Group with basic information.
func NewGroup(name string) *Group {
	return &Group{
		Name:            name,
		HostIDs:         []string{},
		Vars:            make(map[string]string),
		ChildGroupNames: []string{},
	}
}

// Identifiable interface implementation
func (g *Group) GetID() string {
	return g.Name // Groups use name as ID
}

// Nameable interface implementation
func (g *Group) GetName() string {
	return g.Name
}

func (g *Group) SetName(name string) {
	g.Name = name
}

// Describable interface implementation
func (g *Group) GetDescription() string {
	return g.Description
}

func (g *Group) SetDescription(desc string) {
	g.Description = desc
}

// Validate checks if the Group has valid configuration.
func (g *Group) Validate() error {
	if g.Name == "" {
		return fmt.Errorf("group name cannot be empty")
	}
	return nil
}

// Clone creates a deep copy of the Group.
func (g *Group) Clone() interface{} {
	clone := *g
	clone.HostIDs = make([]string, len(g.HostIDs))
	copy(clone.HostIDs, g.HostIDs)
	clone.ChildGroupNames = make([]string, len(g.ChildGroupNames))
	copy(clone.ChildGroupNames, g.ChildGroupNames)
	clone.Vars = make(map[string]string, len(g.Vars))
	for k, v := range g.Vars {
		clone.Vars[k] = v
	}
	return &clone
}

// VarContainer interface implementation
func (g *Group) GetVar(key string) (string, bool) {
	val, ok := g.Vars[key]
	return val, ok
}

func (g *Group) SetVar(key, value string) {
	if g.Vars == nil {
		g.Vars = make(map[string]string)
	}
	g.Vars[key] = value
}

func (g *Group) GetAllVars() map[string]string {
	return g.Vars
}

// AddHost adds a host ID to the group (prevents duplicates).
func (g *Group) AddHost(hostID string) {
	if !g.HasHost(hostID) {
		g.HostIDs = append(g.HostIDs, hostID)
	}
}

// RemoveHost removes a host ID from the group.
func (g *Group) RemoveHost(hostID string) {
	for i, id := range g.HostIDs {
		if id == hostID {
			g.HostIDs = append(g.HostIDs[:i], g.HostIDs[i+1:]...)
			return
		}
	}
}

// HasHost checks if the group contains a specific host ID.
func (g *Group) HasHost(hostID string) bool {
	for _, id := range g.HostIDs {
		if id == hostID {
			return true
		}
	}
	return false
}

// AddChildGroup adds a child group name (prevents duplicates).
func (g *Group) AddChildGroup(groupName string) {
	if !g.HasChildGroup(groupName) {
		g.ChildGroupNames = append(g.ChildGroupNames, groupName)
	}
}

// RemoveChildGroup removes a child group name.
func (g *Group) RemoveChildGroup(groupName string) {
	for i, name := range g.ChildGroupNames {
		if name == groupName {
			g.ChildGroupNames = append(g.ChildGroupNames[:i], g.ChildGroupNames[i+1:]...)
			return
		}
	}
}

// HasChildGroup checks if this group has a specific child group.
func (g *Group) HasChildGroup(groupName string) bool {
	for _, name := range g.ChildGroupNames {
		if name == groupName {
			return true
		}
	}
	return false
}

// HostCount returns the number of hosts in this group (not including child groups).
func (g *Group) HostCount() int {
	return len(g.HostIDs)
}
