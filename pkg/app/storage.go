package app

import (
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitStorage() *minio.Client {
	endpoint := os.Getenv("S3_ENDPOINT") // Correct: only hostname and port
	accessKeyID := os.Getenv("S3_KEY")
	secretAccessKey := os.Getenv("S3_SECRET")
	useSSL := false // Set to true if using HTTPS

	minioClient, _ := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})

	return minioClient
}
