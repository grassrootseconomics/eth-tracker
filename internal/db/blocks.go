package db

import (
	"bytes"

	"github.com/bits-and-blooms/bitset"
	bolt "go.etcd.io/bbolt"
)

func (d *DB) SetLowerBound(v uint64) error {
	return d.setUint64(lowerBoundKey, v)
}

func (d *DB) GetLowerBound() (uint64, error) {
	v, err := d.get(lowerBoundKey)
	if err != nil {
		return 0, err
	}
	return unmarshalUint64(v), nil
}

func (d *DB) SetUpperBound(v uint64) error {
	return d.setUint64(upperBoundKey, v)
}

func (d *DB) GetUpperBound() (uint64, error) {
	v, err := d.get(upperBoundKey)
	if err != nil {
		return 0, err
	}
	return unmarshalUint64(v), nil
}

func (d *DB) SetValue(v uint64) error {
	return d.setUint64AsKey(v)
}

func (d *DB) GetMissingValuesBitSet(lowerBound uint64, upperBound uint64) (*bitset.BitSet, error) {
	var (
		b bitset.BitSet
	)

	err := d.db.View(func(tx *bolt.Tx) error {
		var (
			lowerRaw = marshalUint64(lowerBound)
			upperRaw = marshalUint64(upperBound)
		)

		for i := lowerBound; i <= upperBound; i++ {
			b.Set(uint(i))
		}

		c := tx.Bucket([]byte("blocks")).Cursor()

		for k, _ := c.Seek(lowerRaw); k != nil && bytes.Compare(k, upperRaw) <= 0; k, _ = c.Next() {
			b.Clear(uint(unmarshalUint64(k)))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &b, nil
}

// func (d *DB) Cleanup() error {
// 	var (
// 		safeToDeleteKeys [][]byte
// 	)

// 	err := d.db.View(func(txn *badger.Txn) error {
// 		lowerBound, err := d.get(lowerBoundKey)
// 		if err != nil {
// 			return err
// 		}

// 		lowerBound = marshalUint64(unmarshalUint64(lowerBound) - 1)

// 		opts := badger.DefaultIteratorOptions
// 		opts.PrefetchValues = false

// 		it := txn.NewIterator(opts)
// 		defer it.Close()

// 		for it.Rewind(); it.Valid(); it.Next() {
// 			k := it.Item().Key()

// 			if bytes.Compare(k, lowerBound) > 0 {
// 				return nil
// 			}

// 			safeToDeleteKeys = append(safeToDeleteKeys, it.Item().KeyCopy(nil))
// 		}

// 		return nil
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	wb := d.db.NewWriteBatch()
// 	for _, k := range safeToDeleteKeys {
// 		if err := wb.Delete(k); err != nil {
// 			return nil
// 		}
// 	}

// 	if err := wb.Flush(); err != nil {
// 		return err
// 	}

// 	return nil
// }
