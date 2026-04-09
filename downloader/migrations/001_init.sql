CREATE TABLE IF NOT EXISTS download_requests (
    id BIGSERIAL PRIMARY KEY,
    timeout_seconds INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PROCESS',
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS files (
    id BIGSERIAL PRIMARY KEY,
    request_id BIGINT NOT NULL REFERENCES download_requests(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    file_id BIGINT,
    error_code VARCHAR(20),
    content BYTEA,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);