package archive

import "context"

type ColdStorage interface {
	UploadArchive(ctx context.Context, objectKey string, payload []byte, contentType string) error
	Close() error
}

type noopColdStorage struct{}

func NewNoopColdStorage() ColdStorage {
	return &noopColdStorage{}
}

func (s *noopColdStorage) UploadArchive(context.Context, string, []byte, string) error {
	return nil
}

func (s *noopColdStorage) Close() error {
	return nil
}
