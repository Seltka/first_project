package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/tavanovyt/first_project/downloader/internal/domain"
	"github.com/tavanovyt/first_project/downloader/internal/usecase"
)

type MockRepository struct {
	createDownloadRequestFunc   func(ctx context.Context, req *domain.DownloadRequest) (int64, error)
	createFileFunc              func(ctx context.Context, file *domain.File) (int64, error)
	updateFileAfterDownloadFunc func(ctx context.Context, fileID int64, content []byte, errCode *domain.FileErrorCode) error
	getDownloadRequestFunc      func(ctx context.Context, id int64) (*domain.DownloadRequest, []*domain.File, error)
	getFileContentFunc          func(ctx context.Context, fileID int64) ([]byte, error)
	// ... other methods can be added as needed
}

func (m *MockRepository) CreateDownloadRequest(ctx context.Context, req *domain.DownloadRequest) (int64, error) {
	return m.createDownloadRequestFunc(ctx, req)
}
func (m *MockRepository) CreateFile(ctx context.Context, file *domain.File) (int64, error) {
	return m.createFileFunc(ctx, file)
}
func (m *MockRepository) UpdateFileAfterDownload(ctx context.Context, fileID int64, content []byte, errCode *domain.FileErrorCode) error {
	return m.updateFileAfterDownloadFunc(ctx, fileID, content, errCode)
}
func (m *MockRepository) GetDownloadRequest(ctx context.Context, id int64) (*domain.DownloadRequest, []*domain.File, error) {
	return m.getDownloadRequestFunc(ctx, id)
}
func (m *MockRepository) GetFileContent(ctx context.Context, fileID int64) ([]byte, error) {
	return m.getFileContentFunc(ctx, fileID)
}
func (m *MockRepository) UpdateDownloadRequestStatus(ctx context.Context, id int64, status domain.DownloadStatus) error {
	return nil
}

type MockHTTPDownloader struct {
	downloadFunc func(ctx context.Context, url string) ([]byte, error)
}

func (m *MockHTTPDownloader) Download(ctx context.Context, url string) ([]byte, error) {
	return m.downloadFunc(ctx, url)
}

func TestCreateAsync(t *testing.T) {
	repo := &MockRepository{
		createDownloadRequestFunc: func(ctx context.Context, req *domain.DownloadRequest) (int64, error) {
			return 123, nil
		},
		createFileFunc: func(ctx context.Context, file *domain.File) (int64, error) {
			return 456, nil
		},
		updateFileAfterDownloadFunc: func(ctx context.Context, fileID int64, content []byte, errCode *domain.FileErrorCode) error {
			return nil
		},
	}
	httpClient := &MockHTTPDownloader{
		downloadFunc: func(ctx context.Context, url string) ([]byte, error) {
			return []byte("fake content"), nil
		},
	}

	uc := usecase.NewDownloaderUsecase(repo, httpClient)
	id, err := uc.CreateAsync(context.Background(), []string{"http://example.com"}, 10*time.Second)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 123 {
		t.Errorf("expected id 123, got %d", id)
	}
}
