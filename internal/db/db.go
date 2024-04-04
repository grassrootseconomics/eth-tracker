package db

import (
	"encoding/binary"
	"fmt"
	"log/slog"

	bolt "go.etcd.io/bbolt"
)

type (
	DBOpts struct {
		Logg *slog.Logger
	}

	DB struct {
		db   *bolt.DB
		logg *slog.Logger
	}
)

const (
	dbFolderName = "celo_tracker_blocks_db"

	upperBoundKey = "upper"
	lowerBoundKey = "lower"
)

var (
	sortableOrder = binary.BigEndian
)

func New(o DBOpts) (*DB, error) {
	db, err := bolt.Open(dbFolderName, 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("blocks"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return &DB{
		db:   db,
		logg: o.Logg,
	}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) get(k string) ([]byte, error) {
	var v []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		v = b.Get([]byte(k))
		return nil
	})

	if err != nil {
		return nil, err
	}

	return v, nil
}

func (d *DB) setUint64(k string, v uint64) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		return b.Put([]byte(k), marshalUint64(v))
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) setUint64AsKey(v uint64) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		return b.Put(marshalUint64(v), nil)
	})
	if err != nil {
		return err
	}
	return nil
}

func unmarshalUint64(b []byte) uint64 {
	return sortableOrder.Uint64(b)
}

func marshalUint64(v uint64) []byte {
	b := make([]byte, 8)
	sortableOrder.PutUint64(b, v)
	return b
}
