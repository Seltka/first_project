package domain

import (
	"context"
	"time"
)

type DownloadStatus string

const (
	StatusProcess DownloadStatus = "PROCESS"
	StatusDone    DownloadStatus = "DONE"
)

type FileErrorCode string

const (
	ErrTimeout    FileErrorCode = "TIMEOUT"
	ErrNetwork    FileErrorCode = "NETWORK"
	ErrInvalidURL FileErrorCode = "INVALID_URL"
)

type DownloadRequest struct {
	ID        int64
	Timeout   time.Duration
	Status    DownloadStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type File struct {
	ID        int64
	RequestID int64
	URL       string
	FileID    *int64 // filled after successful download
	ErrorCode *FileErrorCode
	Content   []byte // only used when storing/retrieving
}

type FileInfo struct {
	URL    string
	FileID *int64
	Error  *FileErrorCode
}

// Repository interface (abstract storage)
type Repository interface {
	CreateDownloadRequest(ctx context.Context, req *DownloadRequest) (int64, error)
	UpdateDownloadRequestStatus(ctx context.Context, id int64, status DownloadStatus) error
	GetDownloadRequest(ctx context.Context, id int64) (*DownloadRequest, []*File, error)
	CreateFile(ctx context.Context, file *File) (int64, error)
	GetFileContent(ctx context.Context, fileID int64) ([]byte, error)
	UpdateFileAfterDownload(ctx context.Context, fileID int64, content []byte, errCode *FileErrorCode) error
}

// Downloader interface (external HTTP client)
type HTTPDownloader interface {
	Download(ctx context.Context, url string) ([]byte, error)
}
