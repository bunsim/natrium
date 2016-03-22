package natrium

import (
	"encoding/json"
	"errors"
	"fmt"
)

// #cgo LDFLAGS: -Wl,-Bstatic -lsodium -Wl,-Bdynamic
// #include <stdio.h>
// #include <sodium.h>
import "C"

// EdDSA private key type
type EdDSAPrivate []byte

// EdDSA public key type
type EdDSAPublic []byte

func (k EdDSAPublic) String() string {
	return fmt.Sprintf("dsapub:%x", []byte(k))
}

func (k EdDSAPrivate) String() string {
	return fmt.Sprintf("dsaprv:%x", []byte(k))
}

var EdDSAPublicLength = 0
var EdDSAPrivateLength = 0
var EdDSASignatureLength = 0

// EdDSAGenerateKeys generates an EdDSA private key. The public key
// can be derived from the private key, so there is no issue.
// Keys are represented by byte slices, and can be cast to and from them.
func EdDSAGenerateKey() EdDSAPrivate {
	priv := make([]byte, EdDSAPrivateLength)
	publ := make([]byte, EdDSAPublicLength)
	rv := C.crypto_sign_keypair((*C.uchar)(&publ[0]), (*C.uchar)(&priv[0]))
	if rv != 0 {
		panic("crypto_sign_keypair returned non-zero")
	}
	return priv
}

// PublicKey obtains the public component of an EdDSA private key.
func (priv EdDSAPrivate) PublicKey() EdDSAPublic {
	toret := make([]byte, EdDSAPublicLength)
	rv := C.crypto_sign_ed25519_sk_to_pk((*C.uchar)(&toret[0]),
		(*C.uchar)(&priv[0]))
	if rv != 0 {
		panic("crypto_sign_ed25519_sk_to_pk returned non-zero")
	}
	return toret
}

// Sign signs a message using the given EdDSA private key, returning the signature.
func (priv EdDSAPrivate) Sign(message []byte) []byte {
	signature := make([]byte, EdDSASignatureLength)
	rv := C.crypto_sign_detached(
		(*C.uchar)(&signature[0]),
		nil,
		(*C.uchar)(&message[0]),
		C.ulonglong(len(message)),
		(*C.uchar)(&priv[0]))
	if rv != 0 {
		panic("crypto_sign_detached returned non-zero")
	}
	return signature
}

func (publ EdDSAPublic) MarshalJSON() ([]byte, error) {
	return json.Marshal([]byte(publ))
}

// Verify verifies a signature and a message using a public key. If there is
// a problem, then a non-nil value would be returned. A nil value means
// everything is fine.
func (publ EdDSAPublic) Verify(message []byte, signature []byte) error {
	if len(signature) != EdDSASignatureLength {
		panic(fmt.Sprintf("Signature passed has the wrong length (%v != %v)",
			len(signature), EdDSASignatureLength))
	}
	rv := C.crypto_sign_verify_detached(
		(*C.uchar)(&signature[0]),
		(*C.uchar)(&message[0]),
		C.ulonglong(len(message)),
		(*C.uchar)(&publ[0]))
	if rv != 0 {
		return errors.New("EdDSA signature is forged!")
	}
	return nil
}

func init() {
	C.sodium_init()
	EdDSAPrivateLength = C.crypto_sign_SECRETKEYBYTES
	EdDSAPublicLength = C.crypto_sign_PUBLICKEYBYTES
	EdDSASignatureLength = C.crypto_sign_BYTES
}
