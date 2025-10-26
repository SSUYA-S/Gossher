package inventory

import (
	"fmt"
	"gossher/internal/config"
	"sync"
)

// Manager manages all inventory entities and coordinates their relationships.
type Manager struct {
	basePath string

	credentials map[string]*Credential
	hosts       map[string]*Host
	groups      map[string]*Group

	mu sync.RWMutex // For thread-safe operations
}

// NewManager creates a new InventoryManager.
func NewManager(basePath string) *Manager {
	if basePath == "" {
		basePath = config.Get().DataDir
	}

	return &Manager{
		basePath:    basePath,
		credentials: make(map[string]*Credential),
		hosts:       make(map[string]*Host),
		groups:      make(map[string]*Group),
	}
}

// LoadAll loads all entities from disk into memory.
func (m *Manager) LoadAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load credentials first (hosts may reference them)
	credentials, err := LoadAllCredentials()
	if err != nil {
		return fmt.Errorf("failed to load credentials: %w", err)
	}
	for _, c := range credentials {
		c.SetBasePath(m.basePath)
		m.credentials[c.ID] = c
	}

	// Load hosts
	hosts, err := LoadAllHosts()
	if err != nil {
		return fmt.Errorf("failed to load hosts: %w", err)
	}
	for _, h := range hosts {
		h.SetBasePath(m.basePath)
		m.hosts[h.ID] = h
	}

	// Load groups
	groups, err := LoadAllGroups()
	if err != nil {
		return fmt.Errorf("failed to load groups: %w", err)
	}
	for _, g := range groups {
		g.SetBasePath(m.basePath)
		m.groups[g.Name] = g
	}

	// Validate relationships
	return m.validateRelationships()
}

// validateRelationships checks if all references are valid.
func (m *Manager) validateRelationships() error {
	// Check if hosts reference valid credentials
	for _, host := range m.hosts {
		if host.CredentialID != "" {
			if _, exists := m.credentials[host.CredentialID]; !exists {
				return fmt.Errorf("host %s references non-existent credential: %s",
					host.ID, host.CredentialID)
			}
		}
	}

	// Check if groups reference valid hosts
	for _, group := range m.groups {
		for _, hostID := range group.HostIDs {
			if _, exists := m.hosts[hostID]; !exists {
				return fmt.Errorf("group %s references non-existent host: %s",
					group.Name, hostID)
			}
		}

		// Check if groups reference valid child groups
		for _, childName := range group.ChildGroupNames {
			if _, exists := m.groups[childName]; !exists {
				return fmt.Errorf("group %s references non-existent child group: %s",
					group.Name, childName)
			}
		}
	}

	return nil
}

// ===== Credential Operations =====

// AddCredential adds a credential to memory and saves it to disk.
func (m *Manager) AddCredential(c *Credential) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c == nil {
		return fmt.Errorf("credential cannot be nil")
	}

	c.SetBasePath(m.basePath)

	if err := c.Save(); err != nil {
		return err
	}

	m.credentials[c.ID] = c
	return nil
}

// GetCredential retrieves a credential from memory.
func (m *Manager) GetCredential(id string) (*Credential, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	c, ok := m.credentials[id]
	if !ok {
		return nil, fmt.Errorf("credential %s not found", id)
	}
	return c, nil
}

// UpdateCredential updates a credential in memory and saves it to disk.
func (m *Manager) UpdateCredential(c *Credential) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c == nil {
		return fmt.Errorf("credential cannot be nil")
	}

	c.SetBasePath(m.basePath)

	if err := c.Save(); err != nil {
		return err
	}

	m.credentials[c.ID] = c
	return nil
}

// RemoveCredential removes a credential from memory and disk.
func (m *Manager) RemoveCredential(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if any host is using this credential
	for _, host := range m.hosts {
		if host.CredentialID == id {
			return fmt.Errorf("credential %s is in use by host %s", id, host.ID)
		}
	}

	c, ok := m.credentials[id]
	if !ok {
		return fmt.Errorf("credential %s not found", id)
	}

	if err := c.Delete(); err != nil {
		return err
	}

	delete(m.credentials, id)
	return nil
}

// ListCredentials returns all credentials.
func (m *Manager) ListCredentials() []*Credential {
	m.mu.RLock()
	defer m.mu.RUnlock()

	creds := make([]*Credential, 0, len(m.credentials))
	for _, c := range m.credentials {
		creds = append(creds, c)
	}
	return creds
}

// ===== Host Operations =====

// AddHost adds a host to memory and saves it to disk.
func (m *Manager) AddHost(h *Host) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if h == nil {
		return fmt.Errorf("host cannot be nil")
	}

	// Validate credential reference
	if h.CredentialID != "" {
		if _, exists := m.credentials[h.CredentialID]; !exists {
			return fmt.Errorf("credential %s not found", h.CredentialID)
		}
	}

	h.SetBasePath(m.basePath)

	if err := h.Save(); err != nil {
		return err
	}

	m.hosts[h.ID] = h
	return nil
}

// GetHost retrieves a host from memory.
func (m *Manager) GetHost(id string) (*Host, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	h, ok := m.hosts[id]
	if !ok {
		return nil, fmt.Errorf("host %s not found", id)
	}
	return h, nil
}

// UpdateHost updates a host in memory and saves it to disk.
func (m *Manager) UpdateHost(h *Host) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if h == nil {
		return fmt.Errorf("host cannot be nil")
	}

	// Validate credential reference
	if h.CredentialID != "" {
		if _, exists := m.credentials[h.CredentialID]; !exists {
			return fmt.Errorf("credential %s not found", h.CredentialID)
		}
	}

	h.SetBasePath(m.basePath)

	if err := h.Save(); err != nil {
		return err
	}

	m.hosts[h.ID] = h
	return nil
}

// RemoveHost removes a host from memory and disk.
func (m *Manager) RemoveHost(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	h, ok := m.hosts[id]
	if !ok {
		return fmt.Errorf("host %s not found", id)
	}

	// Remove from all groups
	for _, group := range m.groups {
		if group.HasHost(id) {
			group.RemoveHost(id)
			group.Save() // Save updated group
		}
	}

	if err := h.Delete(); err != nil {
		return err
	}

	delete(m.hosts, id)
	return nil
}

// ListHosts returns all hosts.
func (m *Manager) ListHosts() []*Host {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hosts := make([]*Host, 0, len(m.hosts))
	for _, h := range m.hosts {
		hosts = append(hosts, h)
	}
	return hosts
}

// ===== Group Operations =====

// AddGroup adds a group to memory and saves it to disk.
func (m *Manager) AddGroup(g *Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if g == nil {
		return fmt.Errorf("group cannot be nil")
	}

	// Validate host references
	for _, hostID := range g.HostIDs {
		if _, exists := m.hosts[hostID]; !exists {
			return fmt.Errorf("host %s not found", hostID)
		}
	}

	// Validate child group references
	for _, childName := range g.ChildGroupNames {
		if _, exists := m.groups[childName]; !exists {
			return fmt.Errorf("child group %s not found", childName)
		}
	}

	g.SetBasePath(m.basePath)

	if err := g.Save(); err != nil {
		return err
	}

	m.groups[g.Name] = g
	return nil
}

// GetGroup retrieves a group from memory.
func (m *Manager) GetGroup(name string) (*Group, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	g, ok := m.groups[name]
	if !ok {
		return nil, fmt.Errorf("group %s not found", name)
	}
	return g, nil
}

// UpdateGroup updates a group in memory and saves it to disk.
func (m *Manager) UpdateGroup(g *Group) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if g == nil {
		return fmt.Errorf("group cannot be nil")
	}

	// Validate host references
	for _, hostID := range g.HostIDs {
		if _, exists := m.hosts[hostID]; !exists {
			return fmt.Errorf("host %s not found", hostID)
		}
	}

	g.SetBasePath(m.basePath)

	if err := g.Save(); err != nil {
		return err
	}

	m.groups[g.Name] = g
	return nil
}

// RemoveGroup removes a group from memory and disk.
func (m *Manager) RemoveGroup(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if any group has this as a child
	for _, group := range m.groups {
		if group.HasChildGroup(name) {
			return fmt.Errorf("group %s is a child of group %s", name, group.Name)
		}
	}

	g, ok := m.groups[name]
	if !ok {
		return fmt.Errorf("group %s not found", name)
	}

	if err := g.Delete(); err != nil {
		return err
	}

	delete(m.groups, name)
	return nil
}

// ListGroups returns all groups.
func (m *Manager) ListGroups() []*Group {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groups := make([]*Group, 0, len(m.groups))
	for _, g := range m.groups {
		groups = append(groups, g)
	}
	return groups
}

// ===== Query Methods =====

// GetHostsByGroup returns all hosts in a group (direct members only).
func (m *Manager) GetHostsByGroup(groupName string) ([]*Host, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group, ok := m.groups[groupName]
	if !ok {
		return nil, fmt.Errorf("group %s not found", groupName)
	}

	hosts := make([]*Host, 0, len(group.HostIDs))
	for _, id := range group.HostIDs {
		if h, ok := m.hosts[id]; ok {
			hosts = append(hosts, h)
		}
	}
	return hosts, nil
}

// GetAllHostsInGroup returns all hosts in a group including child groups (recursive).
func (m *Manager) GetAllHostsInGroup(groupName string) ([]*Host, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	group, ok := m.groups[groupName]
	if !ok {
		return nil, fmt.Errorf("group %s not found", groupName)
	}

	hostMap := make(map[string]*Host)

	var collectHosts func(g *Group)
	collectHosts = func(g *Group) {
		// Add direct hosts
		for _, hostID := range g.HostIDs {
			if h, ok := m.hosts[hostID]; ok {
				hostMap[hostID] = h
			}
		}

		// Recursively add from child groups
		for _, childName := range g.ChildGroupNames {
			if childGroup, ok := m.groups[childName]; ok {
				collectHosts(childGroup)
			}
		}
	}

	collectHosts(group)

	hosts := make([]*Host, 0, len(hostMap))
	for _, h := range hostMap {
		hosts = append(hosts, h)
	}
	return hosts, nil
}

// FindHostsByTag returns all hosts with a specific tag.
func (m *Manager) FindHostsByTag(tag string) []*Host {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var hosts []*Host
	for _, h := range m.hosts {
		if h.HasTag(tag) {
			hosts = append(hosts, h)
		}
	}
	return hosts
}

// FindHostsByCredential returns all hosts using a specific credential.
func (m *Manager) FindHostsByCredential(credentialID string) []*Host {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var hosts []*Host
	for _, h := range m.hosts {
		if h.CredentialID == credentialID {
			hosts = append(hosts, h)
		}
	}
	return hosts
}

// GetHostCredential returns the effective credential for a host.
func (m *Manager) GetHostCredential(hostID string) (*Credential, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	host, ok := m.hosts[hostID]
	if !ok {
		return nil, fmt.Errorf("host %s not found", hostID)
	}

	// If using credential reference
	if host.CredentialID != "" {
		if cred, ok := m.credentials[host.CredentialID]; ok {
			return cred, nil
		}
		return nil, fmt.Errorf("credential %s not found", host.CredentialID)
	}

	// Create temporary credential from inline auth
	if host.User != "" {
		return &Credential{
			ID:       fmt.Sprintf("inline-%s", hostID),
			Name:     fmt.Sprintf("Inline auth for %s", host.Name),
			User:     host.User,
			KeyPath:  host.KeyPath,
			Password: host.Password,
		}, nil
	}

	return nil, fmt.Errorf("host %s has no valid authentication", hostID)
}

// ===== Statistics =====

// Stats returns statistics about the inventory.
func (m *Manager) Stats() InventoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return InventoryStats{
		CredentialCount: len(m.credentials),
		HostCount:       len(m.hosts),
		GroupCount:      len(m.groups),
	}
}

// InventoryStats holds statistics.
type InventoryStats struct {
	CredentialCount int
	HostCount       int
	GroupCount      int
}
