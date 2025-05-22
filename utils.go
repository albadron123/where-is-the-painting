package main

import "math/rand"

func generateStr(size int) string {
	ALPHABET_STRING := "abcdefghijklmnopqrstuvwxyz"
	res := make([]byte, size)
	for i := 0; i < size; i++ {
		res[i] = ALPHABET_STRING[rand.Intn(26)]
	}
	return string(res)
}
