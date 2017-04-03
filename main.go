package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

type request struct { // Each request stores the data associated with an image requested from a url tied to an id, so future reuquests will be cached
	id  int        // The id of the request
	url string     // Mostly for detecting cached requests
	img image.RGBA // The image rgba data stored in the request
	mux sync.Mutex // A mutex to stop race conditions
}

var idMax int
var sizeMax int
var requests []request

func main() {
	var host string
	var port string
	flag.StringVar(&host, "h", "localhost", "host to listen on for the webserver")
	flag.StringVar(&port, "p", "8000", "port to listen on for the webserver")
	flag.IntVar(&idMax, "m", 9, "the maximum number of id slots avaliable to cache images")
	flag.IntVar(&sizeMax, "s", 512, "the maximum square size of the image")
	flag.Parse()

	requests = make([]request, idMax)

	http.HandleFunc("/", serve)
	bind := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("listening on %s...\n", bind)
	log.Panic(http.ListenAndServe(bind, nil))
}

func serve(response http.ResponseWriter, request *http.Request) {
	target := request.URL.Query().Get("target")
	segmentString := request.URL.Query().Get("segment")
	idString := request.URL.Query().Get("id")

	id, err := strconv.ParseInt(idString, 10, 0) // Parse the id string into a base 10 Go integer
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	if id < 0 || id > 9 { // Make sure the provided ID isn't out of range
		http.Error(response, "Provided parameter id is out of range (must be 0>id>"+strconv.Itoa(idMax)+")", 400)
		return
	}

	segment, err := strconv.ParseInt(segmentString, 10, 0) // Parse the segment string into a base 10 Go integer
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	u, err := url.Parse(target) // Parse the raw URL value we were given into somthing we can work with
	if err != nil {
		http.Error(response, err.Error(), 400)
		return
	}

	requests[id].mux.Lock()         // Lock the mutex to stop data races (this will block if the mutex is currently in use until it is done)
	defer requests[id].mux.Unlock() // Make sure the mutex gets unlocked

	// Put all relevent data for the new url into the struct
	if requests[id].url != target {
		if !u.IsAbs() { // Is our URL absolute or not?
			u.Scheme = "http"
		} else { // If our URL is absolute, make sure the protocol is http(s)
			if strings.HasPrefix(u.Scheme, "http") {
				u.Scheme = "http"
			}
		}

		targetFormatted := u.String() // Turn our type URL back into a nice easy string, and store it in a variable
		fetchedResponse, err := http.Get(targetFormatted)
		if err != nil {
			http.Error(response, err.Error(), 400)
			return
		}
		fetchedBody, err := ioutil.ReadAll(fetchedResponse.Body) // Read the response into another variable
		if err != nil {
			http.Error(response, err.Error(), 400)
			return
		}

		img, _, err := image.Decode(strings.NewReader(string(fetchedBody))) // Decode the data into an image
		if err != nil {
			http.Error(response, err.Error(), 400)
			return
		}
		m := resize.Thumbnail(uint(sizeMax), uint(sizeMax), img, resize.Lanczos3) // Resize the image using the Lanczos filter
		rect := m.Bounds()                                                        // Get the bounds of the resized image
		rgba := image.NewRGBA(rect)                                               // Make a new RGBA container for our resized image
		draw.Draw(rgba, rect, m, rect.Min, draw.Src)                              // Draw the resize image onto the rgba image

		requests[id].url = target // Set a new url for this id (We do it here because this probably can't fail at this point and valid data should be in)
		requests[id].img = *rgba  // Put the image data into the struct
	}

	// Serialize and return all image data for the requested segment
	if segment < 0 || segment > int64(sizeMax)-1 {
		http.Error(response, "Segment out of range (must be 0>segment>"+strconv.Itoa(sizeMax)+")", 400)
		return
	}

	var output string
	for i := 0; i <= sizeMax-1; i++ {
		if i > requests[id].img.Bounds().Max.X-1 { // If we're outside the bounds of our resized image we're out
			break
		} else if int(segment) > requests[id].img.Bounds().Max.Y { // Make sure we're not out of Y axis bounds
			break
		}
		offset := requests[id].img.PixOffset(i, int(segment))
		values := [3]int{int(requests[id].img.Pix[offset]), int(requests[id].img.Pix[offset+1]), int(requests[id].img.Pix[offset+2])}
		output += strconv.Itoa(values[0]) + "," + strconv.Itoa(values[1]) + "," + strconv.Itoa(values[2]) + " "
	}
	output += "[n]"
	fmt.Fprint(response, output)
}

// detectedContentType := http.DetectContentType(fetchedBody)
// parsedContentType, err := parseContentType(detectedContentType)
// if err != nil {
// 	http.Error(response, err.Error(), 500)
// 	return
// }

// if parsedContentType.Type == "image" {
// 	if parsedContentType.Subtype == "jpeg" || parsedContentType.Subtype == "png" {
// 		serialized, err := serializeImage(strings.NewReader(string(fetchedBody)))
// 		if err != nil {
// 			http.Error(response, err.Error(), 500)
// 			return
// 		}
// 		fmt.Fprint(response, serialized)
// 	}
// }

// func serializeImage(req request, reader io.Reader) (string, error) {
// 	var output string

// 	output += strconv.Itoa(rect.Max.X) + " " + strconv.Itoa(rect.Max.Y) + "|"
// 	for e := 0; e <= rect.Max.Y-1; e++ {
// 		for i := 0; i <= rect.Max.X-1; i++ {
// 			offset := rgba.PixOffset(i, e)
// 			values := [3]int{int(rgba.Pix[offset]), int(rgba.Pix[offset+1]), int(rgba.Pix[offset+2])}
// 			output += strconv.Itoa(values[0]) + "," + strconv.Itoa(values[1]) + "," + strconv.Itoa(values[2]) + " "
// 		}
// 	}
// 	return output, nil
// }

type contentType struct { // The contentType type holds easily usable information that is normally held as a string for indentifying MIME type and character encoding along with other information
	Type       string            // The first part of the MIME type (eg. "text")
	Subtype    string            // The second part of the MIME type (eg. "html")
	Parameters map[string]string // Any extra information (eg. "charset=utf8") represeted as a map
}

func parseContentType(rawcontype string) (*contentType, error) { // Parse a MIME string into a contentType struct
	rawcontype = strings.ToLower(rawcontype)
	var conType contentType
	conType.Parameters = make(map[string]string)
	splitcontype := strings.Split(rawcontype, " ")
	splitcontype[0] = strings.Replace(splitcontype[0], ";", "", -1)
	mimetype := strings.Split(splitcontype[0], "/")
	if len(mimetype) <= 1 {
		return new(contentType), errors.New("contype: malformed content-type MIME type provided")
	}
	if len(splitcontype) > 1 {
		params := strings.Split(splitcontype[1], ";")
		for it := range params {
			splitparams := strings.Split(params[it], "=")
			conType.Parameters[splitparams[0]] = splitparams[1]
		}
	}
	conType.Type = mimetype[0]
	conType.Subtype = mimetype[1]
	return &conType, nil
}
