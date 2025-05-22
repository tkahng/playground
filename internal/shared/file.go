package shared

import "github.com/google/uuid"

type FileDto struct {
	ID           uuid.UUID `json:"id"`
	Disk         string    `db:"disk" json:"disk"`
	Directory    string    `db:"directory" json:"directory"`
	Filename     string    `db:"filename" json:"filename"`
	OriginalName string    `db:"original_name" json:"original_name"`
	Extension    string    `db:"extension" json:"extension"`
	MimeType     string    `db:"mime_type" json:"mime_type"`
	Size         int64     `db:"size" json:"size"`
}
