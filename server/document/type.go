package document

// DocType represents the type of an indexed document.
type DocType int

const (
	Web DocType = iota
	Local
	RemoteFile
)

// String returns the human readable name of the DocType.
func (t DocType) String() string {
	switch t {
	case Web:
		return "web"
	case Local:
		return "local"
	case RemoteFile:
		return "remote"
	default:
		return "unknown"
	}
}
