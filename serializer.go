package main

import "net/http"

// TODO: Fix all of this

//
// import (
// 	"fmt"
// 	"image"
// 	"image/draw"
// 	"io/ioutil"
// 	"net/http"
// 	"net/url"
// 	"strconv"
// 	"strings"
// )
//
func serialize(response http.ResponseWriter, request *http.Request) {
	// 	target := request.URL.Query().Get("target")
	// 	segmentString := request.URL.Query().Get("segment")
	// 	idString := request.URL.Query().Get("id")
	//
	// 	id, err := strconv.ParseInt(idString, 10, 0) // Parse the id string into a base 10 Go integer
	// 	if err != nil {
	// 		http.Error(response, err.Error(), 400)
	// 		return
	// 	}
	//
	// 	if id < 0 || id > 9 { // Make sure the provided ID isn't out of range
	// 		http.Error(response, "Provided parameter id is out of range (must be 0>id>"+strconv.Itoa(cacheSize)+")", 400)
	// 		return
	// 	}
	//
	// 	segment, err := strconv.ParseInt(segmentString, 10, 0) // Parse the segment string into a base 10 Go integer
	// 	if err != nil {
	// 		http.Error(response, err.Error(), 400)
	// 		return
	// 	}
	//
	// 	u, err := url.Parse(target) // Parse the raw URL value we were given into somthing we can work with
	// 	if err != nil {
	// 		http.Error(response, err.Error(), 400)
	// 		return
	// 	}
	//
	// 	requests[id].mux.Lock()         // Lock the mutex to stop data races (this will block if the mutex is currently in use until it is done)
	// 	defer requests[id].mux.Unlock() // Make sure the mutex gets unlocked
	//
	// 	// Put all relevent data for the new url into the struct
	// 	if requests[id].url != target {
	// 		if !u.IsAbs() { // Is our URL absolute or not?
	// 			u.Scheme = "http"
	// 		} else { // If our URL is absolute, make sure the protocol is http(s)
	// 			if strings.HasPrefix(u.Scheme, "http") {
	// 				u.Scheme = "http"
	// 			}
	// 		}
	//
	// 		targetFormatted := u.String() // Turn our type URL back into a nice easy string, and store it in a variable
	// 		fetchedResponse, err := http.Get(targetFormatted)
	// 		if err != nil {
	// 			http.Error(response, err.Error(), 400)
	// 			return
	// 		}
	// 		fetchedBody, err := ioutil.ReadAll(fetchedResponse.Body) // Read the response into another variable
	// 		if err != nil {
	// 			http.Error(response, err.Error(), 400)
	// 			return
	// 		}
	//
	// 		img, _, err := image.Decode(strings.NewReader(string(fetchedBody))) // Decode the data into an image
	// 		if err != nil {
	// 			http.Error(response, err.Error(), 400)
	// 			return
	// 		}
	// 		m := resize.Thumbnail(uint(imageSizeMax), uint(imageSizeMax), img, resize.Lanczos3) // Resize the image using the Lanczos filter
	// 		rect := m.Bounds()                                                                  // Get the bounds of the resized image
	// 		rgba := image.NewRGBA(rect)                                                         // Make a new RGBA container for our resized image
	// 		draw.Draw(rgba, rect, m, rect.Min, draw.Src)                                        // Draw the resize image onto the rgba image
	//
	// 		requests[id].url = target // Set a new url for this id (We do it here because this probably can't fail at this point and valid data should be in)
	// 		requests[id].img = *rgba  // Put the image data into the struct
	// 	}
	//
	// 	// Serialize and return all image data for the requested segment
	// 	if segment < 0 || segment > int64(imageSizeMax)-1 {
	// 		http.Error(response, "Segment out of range (must be 0>segment>"+strconv.Itoa(imageSizeMax)+")", 400)
	// 		return
	// 	}
	//
	// 	var output string
	// 	for i := 0; i <= imageSizeMax-1; i++ {
	// 		if i > requests[id].img.Bounds().Max.X-1 { // If we're outside the bounds of our resized image we're out
	// 			break
	// 		} else if int(segment) > requests[id].img.Bounds().Max.Y { // Make sure we're not out of Y axis bounds
	// 			break
	// 		}
	// 		offset := requests[id].img.PixOffset(i, int(segment))
	// 		values := [3]int{int(requests[id].img.Pix[offset]), int(requests[id].img.Pix[offset+1]), int(requests[id].img.Pix[offset+2])}
	// 		output += strconv.Itoa(values[0]) + "," + strconv.Itoa(values[1]) + "," + strconv.Itoa(values[2]) + ","
	// 	}
	// 	output += "[n]"
	// 	fmt.Fprint(response, output)
}
