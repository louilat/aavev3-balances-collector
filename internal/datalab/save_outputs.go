package datalab

import (
	"encoding/json"
	"strings"

	"github.com/minio/minio-go"
)

func SaveRecords[T any](endpoint string, accessKeyID string, secretAccessKey string, rec []T, bucket, key string) error {
	useSSL := false
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return err
	}
	// fmt.Printf("%v\n", minioClient)

	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	reader := strings.NewReader(string(data))

	_, err = minioClient.PutObject(bucket, key, reader, reader.Size(), minio.PutObjectOptions{ContentType: "text/plain"})
	if err != nil {
		return err
	}
	// fmt.Printf("%v\n", info)

	return nil
}
