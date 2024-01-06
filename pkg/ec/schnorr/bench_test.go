// Copyright 2013-2016 The btcsuite developers
// Copyright (c) 2015-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package schnorr

import (
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/Hubmakerlabs/replicatr/pkg/ec"
	"github.com/Hubmakerlabs/replicatr/pkg/ec/secp"
)

// hexToBytes converts the passed hex string into bytes and will panic if there
// is an error.  This is only provided for the hard-coded constants so errors in
// the source code can be detected. It will only (and must only) be called with
// hard-coded values.
func hexToBytes(s string) []byte {
	b, e := hex.DecodeString(s)
	if e != nil {
		panic("invalid hex in source file: " + s)
	}
	return b
}

// hexToModNScalar converts the passed hex string into a ModNScalar and will
// panic if there is an error.  This is only provided for the hard-coded
// constants so errors in the source code can be detected. It will only (and
// must only) be called with hard-coded values.
func hexToModNScalar(s string) *btcec.ModNScalar {
	b, e := hex.DecodeString(s)
	if e != nil {
		panic("invalid hex in source file: " + s)
	}
	var scalar btcec.ModNScalar
	if overflow := scalar.SetByteSlice(b); overflow {
		panic("hex in source file overflows mod N scalar: " + s)
	}
	return &scalar
}

// hexToFieldVal converts the passed hex string into a FieldVal and will panic
// if there is an error.  This is only provided for the hard-coded constants so
// errors in the source code can be detected. It will only (and must only) be
// called with hard-coded values.
func hexToFieldVal(s string) *btcec.FieldVal {
	b, e := hex.DecodeString(s)
	if e != nil {
		panic("invalid hex in source file: " + s)
	}
	var f btcec.FieldVal
	if overflow := f.SetByteSlice(b); overflow {
		panic("hex in source file overflows mod P: " + s)
	}
	return &f
}

// fromHex converts the passed hex string into a big integer pointer and will
// panic is there is an error.  This is only provided for the hard-coded
// constants so errors in the source code can bet detected. It will only (and
// must only) be called for initialization purposes.
func fromHex(s string) *big.Int {
	if s == "" {
		return big.NewInt(0)
	}
	r, ok := new(big.Int).SetString(s, 16)
	if !ok {
		panic("invalid hex in source file: " + s)
	}
	return r
}

var testOk bool

// BenchmarkSigVerify benchmarks how long it takes the secp256k1 curve to
// verify signatures.
func BenchmarkSigVerify(b *testing.B) {
	// Randomly generated keypair.
	d := hexToModNScalar("9e0699c91ca1e3b7e3c9ba71eb71c89890872be97576010fe593fbf3fd57e66d")

	privKey := secp256k1.NewSecretKey(d)
	pubKey := privKey.PubKey()

	// Double sha256 of []byte{0x01, 0x02, 0x03, 0x04}
	msgHash := sha256.Sum256([]byte("benchmark"))
	sig, e := Sign(privKey, msgHash[:])
	if e != nil {
		b.Fatalf("unable to sign: %v", e)
	}

	if !sig.Verify(msgHash[:], pubKey) {
		b.Errorf("Signature failed to verify")
		return
	}

	var ok bool

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok = sig.Verify(msgHash[:], pubKey)
	}

	testOk = ok
}

// Used to ensure the compiler doesn't optimize away the benchmark.
var (
	testSig *Signature
	testErr error
)

// BenchmarkSign benchmarks how long it takes to sign a message.
func BenchmarkSign(b *testing.B) {
	// Randomly generated keypair.
	d := hexToModNScalar("9e0699c91ca1e3b7e3c9ba71eb71c89890872be97576010fe593fbf3fd57e66d")
	privKey := secp256k1.NewSecretKey(d)

	// blake256 of []byte{0x01, 0x02, 0x03, 0x04}.
	msgHash := hexToBytes("c301ba9de5d6053caad9f5eb46523f007702add2c62fa39de03146a36b8026b7")

	var auxBytes [32]byte
	copy(auxBytes[:], msgHash)
	auxBytes[0] ^= 1

	var (
		sig *Signature
		e error
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sig, e = Sign(
			privKey, msgHash, CustomNonce(auxBytes), FastSign(),
		)
	}

	testSig = sig
	testErr = e
}

// BenchmarkSignRfc6979 benchmarks how long it takes to sign a message.
func BenchmarkSignRfc6979(b *testing.B) {
	// Randomly generated keypair.
	d := hexToModNScalar("9e0699c91ca1e3b7e3c9ba71eb71c89890872be97576010fe593fbf3fd57e66d")
	privKey := secp256k1.NewSecretKey(d)

	// blake256 of []byte{0x01, 0x02, 0x03, 0x04}.
	msgHash := hexToBytes("c301ba9de5d6053caad9f5eb46523f007702add2c62fa39de03146a36b8026b7")

	var (
		sig *Signature
		e error
	)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sig, e = Sign(privKey, msgHash, FastSign())
	}

	testSig = sig
	testErr = e
}
