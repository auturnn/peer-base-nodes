package wallet

import (
	"crypto/x509"
	"encoding/hex"
	"reflect"
	"testing"
)

const (
	testPrivKey = "30770201010420d4d3938cbcc6896b1e78a07d366e151d8cd57e34a31541e3f3f3926654507a71a00a06082a8648ce3d030107a14403420004fd3dea2396c2dd34b82c63fce283f349a789d4ff23d3ca6ae58f3464ebdc54f548f6601748d4eafb77c973dff0efefb8f23db474cbe6da79cbde477d1e523442"
	testPayload = "00a247a0604f3383e6f176a00f5f5ab10806c311e61c757909fe775c96e6ca96"
	testSign    = "83fc24fc0f552361cbb94c931669451fbabc793c287f6fda94a638b2784ab129718e32235382bfee58ef257212f22c9348b2ba631ba3cd34a26fdb5a3b331bdf"
)

func TestInitWallet(t *testing.T) {
	t.Run("Wallet is created", func(t *testing.T) {
		files = fakeLayer{
			fakeHasWalletFile: func() bool {
				t.Log("i have been called")
				return false
			},
		}
		tw := initWallet()
		if reflect.TypeOf(tw) != reflect.TypeOf(&wallet{}) {
			t.Error("New Wallet should return a new wallet instance")
		}
	})
	t.Run("Wallet is restored", func(t *testing.T) {
		files = fakeLayer{
			fakeHasWalletFile: func() bool {
				t.Log("i have been called")
				return true
			},
		}
		w = nil
		tw := initWallet()
		if reflect.TypeOf(tw) != reflect.TypeOf(&wallet{}) {
			t.Error("New Wallet should return a new wallet instance")
		}
	})
}

func makeTestWallet() *wallet {
	w := &wallet{}
	b, _ := hex.DecodeString(testPrivKey)
	key, _ := x509.ParseECPrivateKey(b)
	w.privateKey = key
	w.address = addrFromKey(key)
	return w
}

func TestSign(t *testing.T) {
	s := Sign(testPayload, initWallet().privateKey)
	_, err := hex.DecodeString(s)
	if err != nil {
		t.Errorf("Sign() should return a hex encoded string, got %s", s)
	}
}

func TestVerify(t *testing.T) {
	type testStruct struct {
		input string
		ok    bool
	}
	tests := []testStruct{
		{testPayload, true},
		{"90a247a0604f3383e6f176a00f5f5ab10806c311e61c757909fe775c96e6ca96", false},
	}

	for _, tc := range tests {
		w := makeTestWallet()
		ok := Verify(testSign, tc.input, w.address)
		if ok != tc.ok {
			t.Error("Verify() could not verify testSign and testPayload")
		}
	}
}

func TestRestoreBigInts(t *testing.T) {
	_, _, err := restoreBigInts("error Payload")
	if err == nil {
		t.Error("restoreBigInts() should return error when payload is not hex")
	}
}

func TestGetAddress(t *testing.T) {
	w := &WalletLayer{}
	prk := w.GetPrivKey()
	addr := addrFromKey(prk)
	if reflect.TypeOf(addr) != reflect.TypeOf(w.GetAddress()) {
		t.Error("GetAddress() should return string")
	}
}
