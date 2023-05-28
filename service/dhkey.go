package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"golang.org/x/crypto/curve25519"
	"io"
)

func GeneratePrivateKey() ([32]byte, error) {
	var privateKey [32]byte
	_, err := rand.Read(privateKey[:])
	if err != nil {
		return privateKey, err
	}
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64
	return privateKey, nil
}

func GeneratePublicKey(privateKey *[32]byte) ([32]byte, error) {
	var publicKey [32]byte
	curve25519.ScalarBaseMult(&publicKey, privateKey)
	return publicKey, nil
}

func ComputeSharedSecret(privateKey *[32]byte, publicKey *[32]byte) ([32]byte, error) {
	var sharedSecret [32]byte
	curve25519.ScalarMult(&sharedSecret, privateKey, publicKey)
	return sharedSecret, nil
}

func EncryptWithSharedSecret(plaintext []byte, sharedSecret *[32]byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, fmt.Errorf("plaintext should not be empty")
	}

	block, err := aes.NewCipher(sharedSecret[:])
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}


func test() {
	// Alice 生成密钥对
	alicePrivateKey, err := GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	alicePublicKey, err := GeneratePublicKey(&alicePrivateKey)
	if err != nil {
		panic(err)
	}

	// Bob 生成密钥对
	bobPrivateKey, err := GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	bobPublicKey, err := GeneratePublicKey(&bobPrivateKey)
	if err != nil {
		panic(err)
	}

	// Alice 和 Bob 通过 DH 密钥交换计算共享密钥
	aliceSharedSecret, err := ComputeSharedSecret(&alicePrivateKey, &bobPublicKey)
	if err != nil {
		panic(err)
	}
	bobSharedSecret, err := ComputeSharedSecret(&bobPrivateKey, &alicePublicKey)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Alice 共享密钥: %x\n", aliceSharedSecret)
	fmt.Printf("Bob 共享密钥: %x\n", bobSharedSecret)
}

// pip install cryptography


// from cryptography.hazmat.primitives.asymmetric import x25519
// from cryptography.hazmat.primitives import serialization

// # 生成密钥对
// private_key = x25519.X25519PrivateKey.generate()
// public_key = private_key.public_key()

// # 将公钥序列化为字节串
// public_key_bytes = public_key.public_bytes(
//     encoding=serialization.Encoding.Raw,
//     format=serialization.PublicFormat.Raw
// )

// # 从字节串中加载公钥
// loaded_public_key = x25519.X25519PublicKey.from_public_bytes(public_key_bytes)

// # 计算共享密钥
// shared_key = private_key.exchange(loaded_public_key)

// print("Shared Key:", shared_key.hex())
