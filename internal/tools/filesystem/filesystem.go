package filesystem

import (
	"github.com/c2fo/vfs/v7/backend/s3"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/conf"
)

type FileSystem struct {
	fs  *s3.FileSystem
	cfg conf.StorageConfig
}

func NewFileSystem(cfg conf.StorageConfig) FileSystem {
	bucketAuth := s3.NewFileSystem(s3.WithOptions(s3.Options{
		AccessKeyID:     cfg.ClientId,
		SecretAccessKey: cfg.ClientSecret,
		Region:          cfg.Region,
		Endpoint:        cfg.EndpointUrl,
	}))

	return FileSystem{
		fs:  bucketAuth,
		cfg: cfg,
	}
}

// func extractExtension(name string) string {
// 	primaryDot := strings.LastIndex(name, ".")

// 	if primaryDot == -1 {
// 		return ""
// 	}

// 	// look for secondary extension
// 	secondaryDot := strings.LastIndex(name[:primaryDot], ".")
// 	if secondaryDot >= 0 {
// 		return name[secondaryDot:]
// 	}

// 	return name[primaryDot:]
// }

// func (fs *FileSystem) NewFileFromURL(ctx context.Context, url string) (vfs.File, error) {
// 	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	res, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer res.Body.Close()

// 	if res.StatusCode < 200 || res.StatusCode > 399 {
// 		return nil, fmt.Errorf("failed to download url %s (%d)", url, res.StatusCode)
// 	}

// 	var buf bytes.Buffer

// 	if _, err = io.Copy(&buf, res.Body); err != nil {
// 		return nil, err
// 	}

// 	return fs.NewFileFromBytes(buf.Bytes(), path.Base(url))
// }

type FileDto struct {
	ID           uuid.UUID `json:"id"`
	Size         int64     `json:"size"`
	Extension    string    `json:"extension"`
	MimeType     string    `json:"mime_type"`
	OriginalName string    `json:"original_name"`
}

// func (fs *FileSystem) NewFileFromBytes(b []byte, name string) (vfs.File, error) {
// 	id := uuid.New()
// 	size := len(b)
// 	if size == 0 {
// 		return nil, errors.New("cannot create an empty file")
// 	}
// 	mime := http.DetectContentType(b)
// 	// f := &FileDto{}
// 	a := mimetype.Lookup(mime)
// 	// f.Size = int64(size)
// 	// f.OriginalName = name
// 	// f.Name = normalizeName(f.Reader, f.OriginalName)

// 	return nil, nil
// }
