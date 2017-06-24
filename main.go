package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"sync"
	"time"
)

var softCacheLimit int // TODO: Implement cache clearing
var hardCacheLimit int
var imageSizeMax int
var cache cacheHashTable

type cacheHashTable struct {
	table map[uint64]imageIndex
	mux   sync.RWMutex
}

type imageIndex struct {
	image      image.RGBA
	lastAccess time.Time
}

func main() {
	var host string
	var port string
	var sizeMaxPower int
	flag.StringVar(&host, "host", "localhost", "host to listen on for the webserver")
	flag.StringVar(&port, "port", "8000", "port to listen on for the webserver")
	flag.IntVar(&softCacheLimit, "softcache", 8, "the number of images allowed in the cache before old images get cleared")
	flag.IntVar(&hardCacheLimit, "hardcache", 32, "the maximum number of images allowed in the cache")
	flag.IntVar(&sizeMaxPower, "maxpower", 11, "a number that represents the maximum image size without downsampling as a power of 2 (2^x, this overrides the downsample query parameter)") // 1024 by defualt
	flag.Parse()

	imageSizeMax = 2 ^ sizeMaxPower

	cache.table = make(map[uint64]imageIndex)

	http.HandleFunc("/initalize/", initalize)
	http.HandleFunc("/serialize/", serialize)
	bind := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("listening on %s...\n", bind)
	log.Panic(http.ListenAndServe(bind, nil))
}
