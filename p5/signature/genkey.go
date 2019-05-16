package signature

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)


/**
generate private key and public key file
 */
func GenRsaKey(bits int) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		fmt.Println(err)
	}
	prk := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: prk,
	}
	file, err := os.Create("private.pem")
	if err != nil {
		fmt.Println(err)
	}
	err = pem.Encode(file, block)
	if err != nil {
		fmt.Println(err)
	}


	publicKey := &privateKey.PublicKey
	puk, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		fmt.Println(err)
	}
	block = &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: puk,
	}
	file, err = os.Create("public.pem")
	if err != nil {
		//return err
		fmt.Println(err)
	}
	err = pem.Encode(file, block)
	if err != nil {

		fmt.Println(err)
	}
}