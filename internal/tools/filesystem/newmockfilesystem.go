package filesystem

func NewMockFileSystem() FileSystem {
	return &S3FileSystem{}
}

type S3FileSystemDecorator struct {
	Delegate *S3FileSystem
}
