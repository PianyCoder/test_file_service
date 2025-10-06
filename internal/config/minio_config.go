package config

type MinioConfig struct {
	BaseURL    string `env:"MINIO_BASE_URL"`
	Port       int    `env:"MINIO_PORT"`
	User       string `env:"MINIO_ROOT_USER"`
	Password   string `env:"MINIO_ROOT_PASSWORD"`
	UseSSL     bool   `env:"MINIO_USE_SSL"`
	BucketName string `env:"MINIO_BUCKET_NAME"`
}
