package filesystem

import "github.com/google/uuid"

type FileDto struct {
	ID           uuid.UUID
	StorageKey   string // full bucket path, e.g. "media/uuid.jpg"
	PublicURL    string // stable public URL stored at upload time
	MimeType     string
	Size         int64
	OriginalName string
	Extension    string
	// Legacy fields retained for backward compatibility.
	Disk      string
	Directory string
	Filename  string
}
