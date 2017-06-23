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
	flag.StringVar(&host, "host", "localhost", "host to listen on for the webserver")
	flag.StringVar(&port, "port", "8000", "port to listen on for the webserver")
	flag.IntVar(&softCacheLimit, "softcache", 8, "the number of images allowed in the cache before old images get cleared")
	flag.IntVar(&hardCacheLimit, "hardcache", 32, "the maximum number of images allowed in the cache")
	flag.IntVar(&imageSizeMax, "imagemax", 2048, "the maximum square size of the image (will be automatically downsampled if larger)")
	flag.Parse()

	cache.table = make(map[uint64]imageIndex)

	http.HandleFunc("/initalize/", initalize)
	http.HandleFunc("/serialize/", serialize)
	bind := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("listening on %s...\n", bind)
	log.Panic(http.ListenAndServe(bind, nil))
}
