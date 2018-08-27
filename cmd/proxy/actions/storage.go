package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gomods/athens/pkg/config/env"
	"github.com/gomods/athens/pkg/storage"
	"github.com/gomods/athens/pkg/storage/fs"
	"github.com/gomods/athens/pkg/storage/gcp"
	"github.com/gomods/athens/pkg/storage/mem"
	"github.com/gomods/athens/pkg/storage/minio"
	"github.com/gomods/athens/pkg/storage/mongo"
	"github.com/spf13/afero"
)

// GetStorage returns storage backend based on env configuration
func GetStorage() (storage.Backend, error) {
	storageType := env.StorageTypeWithDefault("memory")
	var storageRoot string
	var err error

	switch storageType {
	case "memory":
		return mem.NewStorage()
	case "mongo":
		connectionString, err := env.MongoConnectionString()
		if err != nil {
			return nil, err
		}

		certPath := env.MongoCertPath()
		return mongo.NewStorageWithCert(connectionString, certPath)
	case "disk":
		storageRoot, err = env.DiskRoot()
		if err != nil {
			return nil, err
		}
		s, err := fs.NewStorage(storageRoot, afero.NewOsFs())
		if err != nil {
			return nil, fmt.Errorf("could not create new storage from os fs (%s)", err)
		}
		return s, nil
	case "minio":
		endpoint, err := env.MinioEndpoint()
		if err != nil {
			return nil, err
		}
		accessKeyID, err := env.MinioAccessKeyID()
		if err != nil {
			return nil, err
		}
		secretAccessKey, err := env.MinioSecretAccessKey()
		if err != nil {
			return nil, err
		}
		bucketName := env.MinioBucketNameWithDefault("gomods")
		useSSL := true
		if useSSLVar := env.MinioSSLWithDefault("yes"); strings.ToLower(useSSLVar) == "no" {
			useSSL = false
		}
		return minio.NewStorage(endpoint, accessKeyID, secretAccessKey, bucketName, useSSL)
	case "gcp":
		return gcp.New(context.Background())
	default:
		return nil, fmt.Errorf("storage type %s is unknown", storageType)
	}
}
