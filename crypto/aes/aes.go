/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// CBCEncrypt performs the operation.
func CBCEncrypt(origData, key, iv []byte) ([]byte, error) {
	if len(iv) == 0 {
		iv = key
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = Pkcs7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, iv[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// CBCDecrypt performs the operation.
func CBCDecrypt(crypted, key, iv []byte) ([]byte, error) {
	if len(iv) == 0 {
		iv = key
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	// CryptBlocks 对非块对齐输入会 panic，先做显式校验
	if len(crypted) == 0 || len(crypted)%blockSize != 0 {
		return nil, errors.New("aes: ciphertext is not a multiple of the block size")
	}
	blockMode := cipher.NewCBCDecrypter(block, iv[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = UnPadding(origData)
	if origData == nil {
		// 填充非法不能静默返回空明文
		return nil, errors.New("aes: invalid padding")
	}
	return origData, nil
}

// Pkcs7Padding returns the result.
func Pkcs7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	// full slice expression 防止 cap>len 时 append 写穿调用方的底层数组
	return append(cipherText[:len(cipherText):len(cipherText)], padText...)
}

// UnPadding returns the result.
func UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return origData
	}
	// Remove the last byte unpadding times
	unPadding := int(origData[length-1])
	//On decrypt unpadding, read the last byte as m and drop m trailing bytes to recover the plaintext
	if unPadding > length || unPadding == 0 {
		return nil
	}
	return origData[:length-unPadding]
}

// Pkcs5Padding returns the result. PKCS#5 is PKCS#7 with a fixed 8-byte block.
func Pkcs5Padding(cipherText []byte) []byte {
	return Pkcs7Padding(cipherText, 8)
}

// ECBEncrypt performs the operation.
func ECBEncrypt(data, key []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := cipher.BlockSize()
	origData := Pkcs7Padding(data, blockSize)
	ecb := NewECBEncrypter(cipher)
	crypted := make([]byte, len(origData))
	ecb.CryptBlocks(crypted, origData)
	return crypted, nil
}

// ECBDecrypt performs the operation.
func ECBDecrypt(crypted, key []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(crypted) == 0 || len(crypted)%cipher.BlockSize() != 0 {
		return nil, errors.New("aes: ciphertext is not a multiple of the block size")
	}
	blockMode := NewECBDecrypter(cipher)
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = UnPadding(origData)
	if origData == nil {
		return nil, errors.New("aes: invalid padding")
	}
	return origData, nil
}

type ecb struct {
	b         cipher.Block
	blockSize int
}

// newECB creates and returns a new instance.
func newECB(b cipher.Block) *ecb {
	return &ecb{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

type ecbEncrypter ecb

// NewECBEncrypter returns a BlockMode which encrypts in electronic code book
// mode, using the given Block.
func NewECBEncrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbEncrypter)(newECB(b))
}

// BlockSize returns the result.
func (x *ecbEncrypter) BlockSize() int { return x.blockSize }

// CryptBlocks performs the operation.
func (x *ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type ecbDecrypter ecb

// NewECBDecrypter returns a BlockMode which decrypts in electronic code book
// mode, using the given Block
func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return (*ecbDecrypter)(newECB(b))
}

// BlockSize returns the result.
func (x *ecbDecrypter) BlockSize() int { return x.blockSize }

// CryptBlocks performs the operation.
func (x *ecbDecrypter) CryptBlocks(dst, src []byte) {
	/*	if len(src)%x.blockSize != 0 {
			panic("crypto/cipher: input not full blocks")
		}
		if len(dst) < len(src) {
			panic("crypto/cipher: output smaller than input")
		}*/
	for len(src) >= x.blockSize {
		x.b.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}
