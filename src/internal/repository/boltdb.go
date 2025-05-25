package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/quaintdev/webshield/src/internal/apperrors"
	"github.com/quaintdev/webshield/src/internal/entity"

	"go.etcd.io/bbolt"
)

type BoltDataStore struct {
	db *bbolt.DB
}

func NewBoltDataStore(dbPath string) (*BoltDataStore, error) {
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("could not open db: %v", err)
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		for _, bucket := range []string{"users", "configs"} {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return fmt.Errorf("could not create bucket %s: %v", bucket, err)
			}
		}
		return nil
	})
	return &BoltDataStore{
		db: db,
	}, nil
}

func (u *BoltDataStore) Close() error {
	return u.db.Close()
}

//Settings repository impl

func (u *BoltDataStore) GetConfig(ctx context.Context, id string) (_ *entity.Settings, _ error) {
	slog.Debug("Reading config from db", "configId", id)
	var config *entity.Settings
	err := u.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("configs"))
		configData := bucket.Get([]byte(id))
		if configData == nil {
			return apperrors.ErrNotFound
		}
		err := json.Unmarshal(configData, &config)
		if err != nil {
			slog.Error("error unmarshalling read config", "id", id)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (u *BoltDataStore) UpdateConfig(ctx context.Context, config *entity.Settings) error {
	return u.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("configs"))
		configData, err := json.Marshal(config)
		if err != nil {
			slog.Error("failing to marshal while saving to db", "error", err)
			return err
		}
		return bucket.Put([]byte(config.ID), configData)
	})
}

func (u *BoltDataStore) DeleteConfig(ctx context.Context, configId string) (_ error) {
	slog.Debug("deleting config from db", "configId", configId)
	return u.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("configs"))
		err := bucket.Delete([]byte(configId))
		if err != nil {
			slog.Error("error deleting config", "id", configId)
			return err
		}
		return nil
	})
}

func (u *BoltDataStore) GetAllConfigs(ctx context.Context) ([]*entity.Settings, error) {
	var configs []*entity.Settings
	err := u.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte("configs"))
		bucket.ForEach(func(k, v []byte) error {
			var config *entity.Settings
			err := json.Unmarshal(v, &config)
			if err != nil {
				return err
			}
			configs = append(configs, config)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return configs, nil
}
