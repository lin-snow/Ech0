package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 加密
func MD5Encrypt(text string) string {
	hash := md5.New()
	hash.Write([]byte(text))
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes)
}
