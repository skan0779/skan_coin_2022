// Package wallet provides wallet functions with private and public key for skancoin
package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/skan0779/skan_coin_2022/utilities"
)

type wallet struct {
	privateKey *ecdsa.PrivateKey
	Address    string
}

const (
	walletName string = "skancoin.wallet"
)

var w *wallet

// check if the user have wallet
func checkWallet() bool {
	_, err := os.Stat(walletName)
	return !os.IsNotExist(err)
}

func createPrivateKey() *ecdsa.PrivateKey {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	utilities.ErrHandling(err)
	return privateKey
}

func savePrivateKey(privateKey *ecdsa.PrivateKey) {
	privateKeyByte, err := x509.MarshalECPrivateKey(privateKey)
	utilities.ErrHandling(err)
	err = os.WriteFile(walletName, privateKeyByte, 0644)
	utilities.ErrHandling(err)
}

func restorePrivateKey() *ecdsa.PrivateKey {
	privateKeyByte, err := os.ReadFile(walletName)
	utilities.ErrHandling(err)
	privateKey, err := x509.ParseECPrivateKey(privateKeyByte)
	utilities.ErrHandling(err)
	return privateKey
}

func createAddress(key *ecdsa.PrivateKey) string {
	address := append(key.X.Bytes(), key.Y.Bytes()...)
	return fmt.Sprintf("%x", address)
}

func Sign(w *wallet, payload string) string {
	// change payload to byte
	payloadByte, err := hex.DecodeString(payload)
	utilities.ErrHandling(err)
	// get signature(=r,s)
	r, s, err := ecdsa.Sign(rand.Reader, w.privateKey, payloadByte)
	utilities.ErrHandling(err)
	signature := append(r.Bytes(), s.Bytes()...)
	return fmt.Sprintf("%x", signature)
}

func restoreBigInt(data string) (*big.Int, *big.Int, error) {
	dataByte, err := hex.DecodeString(data)
	if err != nil {
		return nil, nil, err
	}
	// r,s or x,y
	aByte := dataByte[:len(dataByte)/2]
	bByte := dataByte[len(dataByte)/2:]
	a, b := big.Int{}, big.Int{}
	a.SetBytes(aByte)
	b.SetBytes(bByte)
	return &a, &b, nil
}

func Verify(signature string, payload string, address string) bool {
	r, s, err := restoreBigInt(signature)
	utilities.ErrHandling(err)
	x, y, err := restoreBigInt(address)
	utilities.ErrHandling(err)
	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	payloadByte, err := hex.DecodeString(payload)
	utilities.ErrHandling(err)
	check := ecdsa.Verify(&publicKey, payloadByte, r, s)
	return check
}

func Wallet() *wallet {
	if w == nil {
		w = &wallet{}
		if checkWallet() {
			// yes > restore wallet from file
			key := restorePrivateKey()
			w.privateKey = key
		} else {
			// no > create private key, save it to file
			key := createPrivateKey()
			savePrivateKey(key)
			w.privateKey = key
		}
		w.Address = createAddress(w.privateKey)
	}
	return w
}
