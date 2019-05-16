package signature

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

var privateKey, publicKey []byte

/**
get private key
 */
func GetPrivateKey() []byte {
	var err error
	privateKey, err = ioutil.ReadFile("private.pem")
	if err != nil {s
		fmt.Println(err)
	}

	fmt.Printf("%s\n", privateKey)
	return privateKey
}

/**
get public key
 */
func GetPublicKey() []byte{
	var err error
	publicKey, err = ioutil.ReadFile("public.pem")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s\n", publicKey)
	return publicKey
}

/**
add signature
 */
func RsaSign(data []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	//获取私钥
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
}

/**
verify signature
 */
func RsaSignVer(data []byte, signature []byte) error {
	hashed := sha256.Sum256(data)
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed[:], signature)
}
