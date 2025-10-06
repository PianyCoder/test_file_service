package config

type ServiceConfig struct {
	UploadLimit   int64 `env:"SERVICE_UPLOAD_LIMIT" envDefault:"10"`
	DownloadLimit int64 `env:"SERVICE_DOWNLOAD_LIMIT" envDefault:"10"`
	ListLimit     int64 `env:"SERVICE_LIST_LIMIT" envDefault:"100"`
	ChunkSize     int   `env:"SERVICE_CHUNK_SIZE" envDefault:"1048576"` // 1 MiB
}
