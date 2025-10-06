package config

type MinioConfig struct {
	Endpoint   string `env:"MINIO_BASE_URL" envDefault:"minio"`
	Port       int    `env:"MINIO_PORT" envDefault:"9000"`
	AccessKey  string `env:"MINIO_ROOT_USER" envDefault:"minioadmin"`
	SecretKey  string `env:"MINIO_ROOT_PASSWORD" envDefault:"minioadmin"`
	UseSSL     bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	BucketName string `env:"MINIO_BUCKET_NAME" envDefault:"images-bucket"`
	Location   string `env:"MINIO_LOCATION" envDefault:"us-east-1"`
}
