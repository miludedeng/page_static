package util

import (
	"crypto/md5"
	"encoding/hex"
)

func ContainsBySlice(md5S []string, s string) bool {
	for _, v := range md5S {
		if v == s {
			return true
		}
	}
	return false
}

func RemoveFromSlice(s []string, e string) []string {
	indexS := []int{}
	for i, v := range s {
		if v == e {
			indexS = append(indexS, i)
		}
	}
	for _, v := range indexS {
		s = append(s[:v], s[v+1:]...)
	}
	return s
}

func EncodeUrl(url string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(string(url)))
	cipherEncode := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherEncode)
}
