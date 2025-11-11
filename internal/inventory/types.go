package inventory

// DocumentType represents the type of document stored in YAML.
type DocumentType string

const (
	TypeHost       DocumentType = "host"
	TypeGroup      DocumentType = "group"
	TypeCredential DocumentType = "credential"
	TypeConfig     DocumentType = "config"
)
