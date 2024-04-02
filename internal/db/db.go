package db

import (
	"encoding/binary"
	"log/slog"

	"github.com/dgraph-io/badger/v4"
)

type (
	DBOpts struct {
		Logg *slog.Logger
	}

	DB struct {
		db   *badger.DB
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
	opts := badger.DefaultOptions(dbFolderName)
	opts.Logger = nil
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

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
	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(k))
		if err != nil {
			return err
		}
		v, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return nil, err
	}

	return v, nil
}

func (d *DB) setUint64(k string, v uint64) error {
	err := d.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(k), marshalUint64(v))
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) setUint64AsKey(v uint64) error {
	err := d.db.Update(func(txn *badger.Txn) error {
		return txn.Set(marshalUint64(v), nil)
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
