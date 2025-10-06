package config

type ServiceConfig struct {
	UploadLimit   int64
	DownloadLimit int64
	ListLimit     int64
	ChunkSize     int
}
