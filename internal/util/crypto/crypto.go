// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const (
	AlgoMD5    = "md5"
	AlgoBcrypt = "bcrypt"
)

const MaxPasswordBytes = 72

func MD5Encrypt(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}

func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func CheckPassword(algo, storedHash, plain string) bool {
	if algo == AlgoBcrypt {
		return bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plain)) == nil
	}
	return MD5Encrypt(plain) == storedHash
}

const randomCharset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomString(length int) string {
	if length <= 0 {
		return ""
	}

	const limit = 256 - (256 % len(randomCharset))

	out := make([]byte, length)
	buf := make([]byte, length)
	for filled := 0; filled < length; {
		if _, err := rand.Read(buf); err != nil {
			panic("crypto: secure random source unavailable: " + err.Error())
		}
		for _, v := range buf {
			if int(v) >= limit {
				continue
			}
			out[filled] = randomCharset[int(v)%len(randomCharset)]
			filled++
			if filled == length {
				break
			}
		}
	}
	return string(out)
}
