package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/auturnn/peer-base-nodes/utils"
)

const (
	walletName       string = "kickshaw"
	walletExtentsion string = ".wallet"
)

type WalletLayer struct{}

func (WalletLayer) GetAddress() string {
	return initWallet().address
}

func (WalletLayer) GetPrivKey() *ecdsa.PrivateKey {
	return initWallet().privateKey
}

type wallet struct {
	privateKey *ecdsa.PrivateKey
	address    string
}

var w *wallet

func restoreBigInts(payload string) (*big.Int, *big.Int, error) {
	bytes, err := hex.DecodeString(payload)
	if err != nil {
		return nil, nil, err
	}

	firstHalfBytes := bytes[:len(bytes)/2]
	secondHalfBytes := bytes[len(bytes)/2:]

	bigA, bigB := big.Int{}, big.Int{}
	bigA.SetBytes(firstHalfBytes)
	bigB.SetBytes(secondHalfBytes)

	return &bigA, &bigB, nil
}

func decodeString(payload string) []byte {
	bytes, err := hex.DecodeString(payload)
	utils.HandleError(err)
	return bytes
}

func Sign(payload string, prk *ecdsa.PrivateKey) string {
	payloadBytes := decodeString(payload)
	r, s, err := ecdsa.Sign(rand.Reader, prk, payloadBytes)
	utils.HandleError(err)

	return encodeBigInts(r.Bytes(), s.Bytes())
}

func Verify(sign, payload, addr string) bool {
	r, s, err := restoreBigInts(sign)
	utils.HandleError(err)

	x, y, err := restoreBigInts(addr)
	utils.HandleError(err)

	//not used privateKey
	puK := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	payloadBytes := decodeString(payload)
	return ecdsa.Verify(&puK, payloadBytes, r, s)
}

func encodeBigInts(a, b []byte) string {
	z := append(a, b...)
	return fmt.Sprintf("%x", z)
}

func addrFromKey(key *ecdsa.PrivateKey) string {
	return encodeBigInts(key.X.Bytes(), key.Y.Bytes())
}

func persistKey(key *ecdsa.PrivateKey) {
	bytes, err := x509.MarshalECPrivateKey(key)
	utils.HandleError(err)
	utils.HandleError(files.writeFile(getWalletPath(), bytes, 0644))
}

func createPrivKey() *ecdsa.PrivateKey {
	prk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	utils.HandleError(err)
	return prk
}

func getWalletPath() string {
	return fmt.Sprintf("./%s_%d%s", walletName, 8080, walletExtentsion)
}

func restoreKey() *ecdsa.PrivateKey {
	keyAsBytes, err := files.readFile(getWalletPath())
	utils.HandleError(err)

	key, err := x509.ParseECPrivateKey(keyAsBytes)
	utils.HandleError(err)

	return key
}

func initWallet() *wallet {
	if w == nil {
		w = &wallet{}
		if files.hasWalletFile() {
			// true = restore from file
			w.privateKey = restoreKey()
		} else {
			//has a wallet already
			// false = create prk , save to file
			key := createPrivKey()
			persistKey(key)
			w.privateKey = key
		}
		w.address = addrFromKey(w.privateKey)
	}
	return w
}
