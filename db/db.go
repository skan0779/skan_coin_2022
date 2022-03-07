// Package db provides databased functions for skancoin
package db

import (
	"fmt"
	"os"

	"github.com/skan0779/skan_coin_2022/utilities"
	bolt "go.etcd.io/bbolt"
)

const (
	dbName        = "blockchain"
	bucketBlocks  = "blocks"
	bucketData    = "data"
	bucketDataKey = "alignment"
)

var db *bolt.DB

func DB() *bolt.DB {
	if db == nil {
		// 1. init db
		db2, err := bolt.Open(getDBName(), 0600, nil)
		utilities.ErrHandling(err)
		db = db2
		// 2. bucket(2) check and create
		err = db.Update(func(t *bolt.Tx) error {
			_, err := t.CreateBucketIfNotExists([]byte(bucketBlocks))
			utilities.ErrHandling(err)
			_, err = t.CreateBucketIfNotExists([]byte(bucketData))
			return err
		})
		utilities.ErrHandling(err)
	}
	return db
}

func Close() {
	DB().Close()
}

func SaveBlock(hash string, data []byte) {

	fmt.Printf("\n Saved Block: %s \n", hash)
	err := DB().Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(bucketBlocks))
		err := b.Put([]byte(hash), data)
		return err
	})
	utilities.ErrHandling(err)
}

func SaveBlockchain(data []byte) {

	err := DB().Update(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(bucketData))
		err := b.Put([]byte(bucketDataKey), data)
		return err
	})
	utilities.ErrHandling(err)
}

func GetBucketData() []byte {
	var data []byte
	DB().View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(bucketData))
		data = b.Get([]byte(bucketDataKey))
		return nil
	})
	return data
}

func GetBucketBlocks(hash string) []byte {
	var data []byte
	DB().View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(bucketBlocks))
		data = b.Get([]byte(hash))
		return nil
	})
	return data
}

func getDBName() string {
	port := os.Args[2][6:]
	return fmt.Sprintf("%s_%s.db", dbName, port)
}

func UpdateBucketBlocks() {
	DB().Update(func(t *bolt.Tx) error {
		err := t.DeleteBucket([]byte(bucketBlocks))
		utilities.ErrHandling(err)
		_, err = t.CreateBucket([]byte(bucketBlocks))
		utilities.ErrHandling(err)
		return nil
	})
}
