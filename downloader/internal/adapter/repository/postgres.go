package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/tavanovyt/first_project/downloader/internal/domain"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

// CreateDownloadRequest inserts a new download job and returns its ID
func (r *PostgresRepo) CreateDownloadRequest(ctx context.Context, req *domain.DownloadRequest) (int64, error) {
	var id int64
	query := `INSERT INTO download_requests (timeout_seconds, status, created_at, updated_at)
              VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRowContext(ctx, query,
		int(req.Timeout.Seconds()),
		req.Status,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&id)
	return id, err
}

// UpdateDownloadRequestStatus changes the status of a job (PROCESS -> DONE)
func (r *PostgresRepo) UpdateDownloadRequestStatus(ctx context.Context, id int64, status domain.DownloadStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE download_requests SET status=$1, updated_at=$2 WHERE id=$3`,
		status, time.Now().UTC(), id)
	return err
}

// GetDownloadRequest returns the job and all its associated files
func (r *PostgresRepo) GetDownloadRequest(ctx context.Context, id int64) (*domain.DownloadRequest, []*domain.File, error) {
	// Fetch the request row
	var req domain.DownloadRequest
	var timeoutSec int
	err := r.db.QueryRowContext(ctx,
		`SELECT id, timeout_seconds, status, created_at, updated_at FROM download_requests WHERE id=$1`, id).
		Scan(&req.ID, &timeoutSec, &req.Status, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		return nil, nil, err
	}
	req.Timeout = time.Duration(timeoutSec) * time.Second

	// Fetch all files for this request
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, request_id, url, file_id, error_code FROM files WHERE request_id=$1`, id)
	if err != nil {
		return &req, nil, err
	}
	defer rows.Close()

	var files []*domain.File
	for rows.Next() {
		f := &domain.File{RequestID: id}
		var fileID sql.NullInt64
		var errorCode sql.NullString
		if err := rows.Scan(&f.ID, &f.RequestID, &f.URL, &fileID, &errorCode); err != nil {
			return &req, nil, err
		}
		if fileID.Valid {
			f.FileID = &fileID.Int64
		}
		if errorCode.Valid {
			code := domain.FileErrorCode(errorCode.String)
			f.ErrorCode = &code
		}
		files = append(files, f)
	}
	return &req, files, nil
}

// CreateFile creates a placeholder entry for a file before download starts
func (r *PostgresRepo) CreateFile(ctx context.Context, file *domain.File) (int64, error) {
	var id int64
	query := `INSERT INTO files (request_id, url) VALUES ($1, $2) RETURNING id`
	err := r.db.QueryRowContext(ctx, query, file.RequestID, file.URL).Scan(&id)
	return id, err
}

// GetFileContent returns the raw bytes of a successfully downloaded file
func (r *PostgresRepo) GetFileContent(ctx context.Context, fileID int64) ([]byte, error) {
	var content []byte
	err := r.db.QueryRowContext(ctx, `SELECT content FROM files WHERE id=$1`, fileID).Scan(&content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// UpdateFileAfterDownload stores the result of a download (content or error)
func (r *PostgresRepo) UpdateFileAfterDownload(ctx context.Context, fileID int64, content []byte, errCode *domain.FileErrorCode) error {
	if errCode != nil {
		// Failed download: store error code, no content
		_, err := r.db.ExecContext(ctx,
			`UPDATE files SET error_code=$1, updated_at=$2 WHERE id=$3`,
			string(*errCode), time.Now().UTC(), fileID)
		return err
	}
	// Successful download: store content and set file_id = id (self reference)
	_, err := r.db.ExecContext(ctx,
		`UPDATE files SET content=$1, file_id=$2, updated_at=$3 WHERE id=$4`,
		content, fileID, time.Now().UTC(), fileID)
	return err
}
