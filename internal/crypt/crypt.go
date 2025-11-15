package crypt

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/hmac"
    "crypto/sha256"
    "io"
)

func HKDF(master []byte, salt []byte, info []byte, length int) []byte {
    prk := hmac.New(sha256.New, salt)
    prk.Write(master)
    prkSum := prk.Sum(nil)
    res := make([]byte, 0, length)
    var t []byte
    var ctr byte = 1
    for len(res) < length {
        h := hmac.New(sha256.New, prkSum)
        h.Write(t)
        h.Write(info)
        h.Write([]byte{ctr})
        t = h.Sum(nil)
        res = append(res, t...)
        ctr++
    }
    return res[:length]
}

func EncryptAEAD(key []byte, nonce []byte, plaintext []byte, aad []byte) ([]byte, error) {
    blk, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    aead, err := cipher.NewGCM(blk)
    if err != nil {
        return nil, err
    }
    ct := aead.Seal(nil, nonce, plaintext, aad)
    return ct, nil
}

func DecryptAEAD(key []byte, nonce []byte, ciphertext []byte, aad []byte) ([]byte, error) {
    blk, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }
    aead, err := cipher.NewGCM(blk)
    if err != nil {
        return nil, err
    }
    pt, err := aead.Open(nil, nonce, ciphertext, aad)
    if err != nil {
        return nil, err
    }
    return pt, nil
}

func RandBytes(r io.Reader, n int) ([]byte, error) {
    b := make([]byte, n)
    _, err := io.ReadFull(r, b)
    return b, err
}