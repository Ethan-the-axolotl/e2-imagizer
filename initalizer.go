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

	"github.com/nfnt/resize"
)

func initalize(response http.ResponseWriter, request *http.Request) {
	target := request.URL.Query().Get("target")
	u, err := url.Parse(target) // Parse the raw URL value we were given into somthing we can work with
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	// Is our URL absolute or not?
	if !u.IsAbs() {
		// The user probably forgot the URL scheme (http, ftp, etc..), we'll try guessing http (we want this to work)
		u.Scheme = "http"
	}

	// If the URL doesn't have a scheme of http(s), make it that way
	if !strings.HasPrefix(u.Scheme, "http") {
		u.Scheme = "http" // This is just a guess, we're hoping that the user just made a typo
	}

	// Make a HTTP request for our URL
	fetchedResponse, err := http.Get(u.String())
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
	requests.mux.RLock()
	if _, exists := requests.table[checksum]; exists {
		// Cache hit! Return what all the relavant information
		requests.mux.RUnlock()
		fmt.Fprint(response, "OK@"+strconv.Itoa(int(checksum))+"@HIT") // e2 doesn't give you access to HTTP status codes (which is silly), so we have to do this
		return
	}
	requests.mux.RUnlock()

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
	requests.mux.Lock()
	requests.table[checksum] = *rgba
	requests.mux.Unlock()

	fmt.Fprint(response, "OK@"+strconv.Itoa(int(checksum))) // e2 doesn't give you access to HTTP status codes (which is silly), so we have to do this
	return
}
