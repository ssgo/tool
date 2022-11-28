package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/tjfoc/gmsm/sm4"
	"strconv"
)

func xor(in, iv []byte) (out []byte) {
	if len(in) != len(iv) {
		return nil
	}

	out = make([]byte, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = in[i] ^ iv[i]
	}
	return
}

func pkcs7Padding(src []byte) []byte {
	padding := sm4.BlockSize - len(src)%sm4.BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func pkcs7UnPadding(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])
	if unpadding > sm4.BlockSize || unpadding == 0 {
		return nil, errors.New("Invalid pkcs7 padding (unpadding > BlockSize || unpadding == 0)")
	}

	pad := src[len(src)-unpadding:]
	for i := 0; i < unpadding; i++ {
		if pad[i] != byte(unpadding) {
			return nil, errors.New("Invalid pkcs7 padding (pad[i] != unpadding)")
		}
	}

	return src[:(length - unpadding)], nil
}

func sm4Cbc(key, iv []byte, in []byte, mode bool) (out []byte, err error) {
	if len(key) < sm4.BlockSize || len(iv) < sm4.BlockSize {
		return nil, errors.New("SM4: invalid key size " + strconv.Itoa(len(key)))
	}
	if len(key) > sm4.BlockSize {
		key = key[0:sm4.BlockSize]
	}
	if len(iv) > sm4.BlockSize {
		iv = iv[0:sm4.BlockSize]
	}
	var inData []byte
	if mode {
		inData = pkcs7Padding(in)
	} else {
		inData = in
	}

	out = make([]byte, len(inData))
	c, err := sm4.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if mode {
		for i := 0; i < len(inData)/16; i++ {
			in_tmp := xor(inData[i*16:i*16+16], iv)
			out_tmp := make([]byte, 16)
			c.Encrypt(out_tmp, in_tmp)
			copy(out[i*16:i*16+16], out_tmp)
			iv = out_tmp
		}
	} else {
		for i := 0; i < len(inData)/16; i++ {
			in_tmp := inData[i*16 : i*16+16]
			out_tmp := make([]byte, 16)
			c.Decrypt(out_tmp, in_tmp)
			out_tmp = xor(out_tmp, iv)
			copy(out[i*16:i*16+16], out_tmp)
			iv = in_tmp
		}
		out, _ = pkcs7UnPadding(out)
	}

	return out, nil
}

func EncryptSM4(in []byte, key, iv []byte) string {
	return base64.URLEncoding.EncodeToString(EncryptSM4Bytes(in, key, iv))
}

func EncryptSM4Bytes(in []byte, key, iv []byte) []byte {
	if enBytes, err := sm4Cbc(key, iv, in, true); err == nil {
		return enBytes
	}
	return in
}

func DecryptSM4(in string, key, iv []byte) []byte {
	if enBytes, err := base64.URLEncoding.DecodeString(in); err == nil {
		return DecryptSM4Bytes(enBytes, key, iv)
	} else {
		return DecryptSM4Bytes([]byte(in), key, iv)
	}
}

func DecryptSM4Bytes(in []byte, key, iv []byte) []byte {
	if deBytes, err := sm4Cbc(key, iv, in, false); err == nil {
		return deBytes
	}
	return in
}
