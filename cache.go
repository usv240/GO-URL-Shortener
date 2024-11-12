package main

import (
	"sync"
)

var cache = make(map[string]string)
var cacheMu sync.Mutex

func cacheURL(shortURL, longURL string) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache[shortURL] = longURL
}

func getURLFromCache(shortURL string) (string, bool) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	longURL, exists := cache[shortURL]
	return longURL, exists
}
