package inventory

// Identifiable represents any entity that has a unique identifier.
type Identifiable interface {
	GetID() string
}

// Nameable represents any entity that has a human-readable name.
type Nameable interface {
	GetName() string
	SetName(name string)
}

// Describable represents any entity that can have a description.
type Describable interface {
	GetDescription() string
	SetDescription(desc string)
}

// Validatable represents any entity that can validate itself.
type Validatable interface {
	Validate() error
}

// Cloneable represents any entity that can create a deep copy of itself.
type Cloneable interface {
	Clone() interface{}
}

// VarContainer represents any entity that can store custom variables.
type VarContainer interface {
	GetVar(key string) (string, bool)
	SetVar(key, value string)
	GetAllVars() map[string]string
}

// Entity combines common interfaces for inventory entities.
type Entity interface {
	Identifiable
	Nameable
	Validatable
	Cloneable
}

// TaggedEntity represents entities that can be tagged.
type TaggedEntity interface {
	Entity
	HasTag(tag string) bool
	AddTag(tag string)
	RemoveTag(tag string)
	GetTags() []string
}
