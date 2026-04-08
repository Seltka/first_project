package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/tavanovyt/first_project/downloader/internal/domain"
	"github.com/tavanovyt/first_project/downloader/internal/usecase"
)

type Handler struct {
	uc *usecase.DownloaderUsecase
}

func NewHandler(uc *usecase.DownloaderUsecase) *Handler {
	return &Handler{uc: uc}
}

// Request body for POST /downloads
type createReq struct {
	Files   []struct{ URL string } `json:"files"`
	Timeout string                 `json:"timeout"`
}

// Response for POST /downloads
type createResp struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

// Response for GET /downloads/{id}
type getStatusResp struct {
	ID     int64          `json:"id"`
	Status string         `json:"status"`
	Files  []fileInfoResp `json:"files"`
}

type fileInfoResp struct {
	URL    string                 `json:"url"`
	FileID *int64                 `json:"file_id,omitempty"`
	Error  *struct{ Code string } `json:"error,omitempty"`
}

// POST /downloads
func (h *Handler) CreateDownload(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if len(req.Files) == 0 {
		http.Error(w, "at least one file required", http.StatusBadRequest)
		return
	}

	timeout, err := time.ParseDuration(req.Timeout)
	if err != nil || timeout <= 0 {
		http.Error(w, "invalid timeout (e.g., '60s')", http.StatusBadRequest)
		return
	}

	urls := make([]string, len(req.Files))
	for i, f := range req.Files {
		urls[i] = f.URL
	}

	id, err := h.uc.CreateAsync(r.Context(), urls, timeout)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := createResp{
		ID:     id,
		Status: string(domain.StatusProcess),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GET /downloads/{id}
func (h *Handler) GetDownloadStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	req, infos, err := h.uc.GetDownloadStatus(r.Context(), id)
	if err != nil {
		http.Error(w, "download request not found", http.StatusNotFound)
		return
	}

	resp := getStatusResp{
		ID:     req.ID,
		Status: string(req.Status),
		Files:  make([]fileInfoResp, len(infos)),
	}
	for i, info := range infos {
		resp.Files[i] = fileInfoResp{URL: info.URL}
		if info.FileID != nil {
			resp.Files[i].FileID = info.FileID
		}
		if info.Error != nil {
			resp.Files[i].Error = &struct{ Code string }{Code: string(*info.Error)}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GET /downloads/{id}/files/{file_id}
func (h *Handler) GetFile(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.PathValue("file_id")
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}

	content, err := h.uc.GetFile(r.Context(), fileID)
	if err != nil {
		http.Error(w, "file not found or not yet downloaded", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(content)
}
