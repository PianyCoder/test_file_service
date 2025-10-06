package config

type ServiceConfig struct {
	UploadLimit   int64 `env:"SERVICE_UPLOAD_LIMIT"`
	DownloadLimit int64 `env:"SERVICE_DOWNLOAD_LIMIT"`
	ListLimit     int64 `env:"SERVICE_LIST_LIMIT"`
	ChunkSize     int   `env:"SERVICE_CHUNK_SIZE"`
}
