package usecase

import (
	"context"
	"time"

	"github.com/tavanovyt/first_project/downloader/internal/domain"
	"golang.org/x/sync/errgroup"
)

type DownloaderUsecase struct {
	repo           domain.Repository
	httpDownloader domain.HTTPDownloader
}

func NewDownloaderUsecase(repo domain.Repository, httpDownloader domain.HTTPDownloader) *DownloaderUsecase {
	return &DownloaderUsecase{
		repo:           repo,
		httpDownloader: httpDownloader,
	}
}

// CreateAsync creates a download request and starts async downloads for all URLs
func (uc *DownloaderUsecase) CreateAsync(ctx context.Context, urls []string, timeout time.Duration) (int64, error) {
	// Create the download request
	req := &domain.DownloadRequest{
		Timeout:   timeout,
		Status:    domain.StatusProcess,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	id, err := uc.repo.CreateDownloadRequest(ctx, req)
	if err != nil {
		return 0, err
	}

	// Create file records for each URL
	fileIDs := make([]int64, len(urls))
	for i, url := range urls {
		file := &domain.File{
			RequestID: id,
			URL:       url,
		}
		fileID, err := uc.repo.CreateFile(ctx, file)
		if err != nil {
			return 0, err
		}
		fileIDs[i] = fileID
	}

	// Start async downloads in background
	go uc.downloadAsync(context.Background(), id, urls, fileIDs, timeout)

	return id, nil
}

// downloadAsync performs the actual downloads concurrently
func (uc *DownloaderUsecase) downloadAsync(ctx context.Context, requestID int64, urls []string, fileIDs []int64, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	for i, url := range urls {
		i, url := i, url // Capture loop variables
		g.Go(func() error {
			data, err := uc.httpDownloader.Download(ctx, url)

			var errCode *domain.FileErrorCode
			if err != nil {
				code := domain.ErrNetwork
				errCode = &code
			}

			// Update the file record with content/error
			return uc.repo.UpdateFileAfterDownload(ctx, fileIDs[i], data, errCode)
		})
	}

	// Wait for all downloads to complete
	_ = g.Wait() // Ignore error, already recorded in DB

	// Update request status to DONE
	_ = uc.repo.UpdateDownloadRequestStatus(context.Background(), requestID, domain.StatusDone)
}

// GetDownloadStatus returns the download request and file info
func (uc *DownloaderUsecase) GetDownloadStatus(ctx context.Context, id int64) (*domain.DownloadRequest, []*domain.FileInfo, error) {
	req, files, err := uc.repo.GetDownloadRequest(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	infos := make([]*domain.FileInfo, len(files))
	for i, file := range files {
		infos[i] = &domain.FileInfo{
			URL:    file.URL,
			FileID: file.FileID,
			Error:  file.ErrorCode,
		}
	}

	return req, infos, nil
}

// GetFile retrieves the content of a downloaded file
func (uc *DownloaderUsecase) GetFile(ctx context.Context, fileID int64) ([]byte, error) {
	return uc.repo.GetFileContent(ctx, fileID)
}
