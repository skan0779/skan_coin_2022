// Package utilities provides utility functions for skancoin
package utilities

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// Shortcut to handle the errors: exit the program if error occurs
func ErrHandling(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Encode any arguement to []byte type
func ToByte(i interface{}) []byte {
	// create buffer: can read/write bytes
	// create encoder
	// encode the data (b= whole block)
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(i)
	ErrHandling(err)
	return buffer.Bytes()
}

// Decode any arguement into i
func FromByte(i interface{}, data []byte) {
	dec := gob.NewDecoder(bytes.NewReader(data))
	ErrHandling(dec.Decode(i))
}

// hashing the input data: string -> []byte()
func Hash(i interface{}) string {
	// "%v" is the default format
	data := fmt.Sprintf("%v", i)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

func Spliter(s string, sep string, i int) string {
	word := strings.Split(s, sep)
	if len(word)-1 < i {
		return ""
	}
	return word[i]
}

func Json(i interface{}) []byte {
	j, err := json.Marshal(i)
	ErrHandling(err)
	return j
}
