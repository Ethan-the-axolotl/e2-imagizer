package main

import (
	"fmt"
	"hash/crc64"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nfnt/resize"
)

func initalize(response http.ResponseWriter, request *http.Request) {
	var err error

	if len(cache.table) >= hardCacheLimit {
		http.Error(response, "Hard cache limit of "+strconv.Itoa(hardCacheLimit)+" hit, please try again later or raise the cache limit.", 500)
	}

	urlTargetString := request.URL.Query().Get("target")
	if urlTargetString == "" {
		http.Error(response, "target parameter cannot be empty.", 400)
	}

	urlTarget, err := url.Parse(urlTargetString) // Parse the raw URL value we were given into somthing we can work with
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	// Is our URL absolute or not?
	if !urlTarget.IsAbs() {
		// The user probably forgot the URL scheme (http, ftp, etc..), we'll try guessing http (we want this to work)
		urlTarget.Scheme = "http"
	}

	// If the URL doesn't have a scheme of http(s), make it that way
	if !strings.HasPrefix(urlTarget.Scheme, "http") {
		urlTarget.Scheme = "http" // This is just a guess, we're hoping that the user just made a typo
	}

	// Make a HTTP request for our URL
	fetchedResponse, err := http.Get(urlTarget.String())
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	// Read the response into a variable
	fetchedBody, err := ioutil.ReadAll(fetchedResponse.Body)
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	detectedConTypeString := http.DetectContentType(fetchedBody)
	detectedConType, err := parseContentType(detectedConTypeString)
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	// Make sure that we're parsing an image
	if detectedConType.Type != "image" {
		http.Error(response, "Invalid response from requested URL, response is of type '"+detectedConType.Type+"' not a image.", 400)
		return
	}

	// Make sure that we support the image we're parsing
	switch detectedConType.Subtype { // govet yells at you if you do this using an if statement... Is this really any better?
	case "jpeg", "gif", "png":
	default:
		http.Error(response, detectedConTypeString+" is unspported, it must be one of the following: image/jpeg, image/gif, image/png.", 400)
		return
	}

	checksum := crc64.Checksum(fetchedBody, crc64.MakeTable(crc64.ECMA)) // Weakly hash the image so that we have something to address it with

	// Check for a cache hit
	cache.mux.RLock()
	if _, exists := cache.table[checksum]; exists {
		// Cache hit! Return what all the relavant information
		cache.mux.RUnlock()
		fmt.Fprint(response, "OK "+strconv.FormatUint(checksum, 10)+" HIT") // e2 doesn't give you access to HTTP status codes (which is silly), so we have to do this
		return
	}
	cache.mux.RUnlock()

	// Try to prune the cache
	cache.mux.Lock()
	for k := range cache.table {
		// Prune things older than an hour
		if time.Now().Sub(cache.table[k].lastAccess) >= time.Hour {
			delete(cache.table, k)
		}
	}
	cache.mux.Unlock()

	// Decode the data into an image
	img, _, err := image.Decode(strings.NewReader(string(fetchedBody)))
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	// Resize the image using the Lanczos filter and draw it into an RGBA
	resizedImage := resize.Thumbnail(uint(imageSizeMax), uint(imageSizeMax), img, resize.Lanczos3) // Resize the images
	rect := resizedImage.Bounds()                                                                  // Get the bounds of the resized image
	rgba := image.NewRGBA(rect)                                                                    // Make a new RGBA container for our resized image
	draw.Draw(rgba, rect, resizedImage, rect.Min, draw.Src)                                        // Draw the resize image onto the rgba image

	// Populate the cache
	cache.mux.Lock()
	tmp := cache.table[checksum]
	tmp.image = *rgba
	tmp.lastAccess = time.Now() // This is for cache-clearing purposes
	cache.table[checksum] = tmp
	cache.mux.Unlock()

	fmt.Fprint(response, "OK "+strconv.FormatUint(checksum, 10)) // e2 doesn't give you access to HTTP status codes (which is silly), so we have to do this
	return
}
