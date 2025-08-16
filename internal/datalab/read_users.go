package datalab

import (
	"encoding/json"
	"io"

	"github.com/minio/minio-go"
)

func ReadActiveUsers(endpoint, bucket, key, accessKeyID, secretAccessKey string) ([]string, error) {
	useSSL := true

	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		return make([]string, 0), err
	}

	usr, err := minioClient.GetObject(bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return make([]string, 0), err
	}

	usersBytes, err := io.ReadAll(usr)
	if err != nil {
		return make([]string, 0), err
	}

	users := make([]string, 0)
	err = json.Unmarshal(usersBytes, &users)
	if err != nil {
		return make([]string, 0), err
	}
	return users, nil
}
