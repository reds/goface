package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func getAListFromAUrl(url string) (list []string) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	r := bufio.NewReader(resp.Body)
	for {
		l, err := r.ReadString('\n')
		if err != nil {
			break
		}
		list = append(list, strings.TrimSpace(l))
	}
	return
}

func faceDetectImages(apiUrl string, imageUrls []string) string {
	body := strings.NewReader(strings.Join(imageUrls, "\n"))
	resp, err := http.Post(apiUrl, "text/plain", body)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(rbody)
}

func faceDetectAnImage(apiUrl, imageUrl string) string {
	url, err := url.Parse(apiUrl)
	q := url.Query()
	q.Set("url", imageUrl)
	url.RawQuery = q.Encode()
	resp, err := http.Get(url.String())
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

type Rect struct {
	X, Y, W, H int
}

type Fd struct {
	Num         int           `json:"num"`
	Url         string        `json:"url"`
	ErrorNum    int           `json:"errorNum"`
	ErrorString string        `json:"errorString,omitempty"`
	NumFaces    int           `json:"numFaces,omitempty"`
	Faces       []Rect        `json:"faces,omitempty"`
	Time        time.Duration `json:"time,omitempty"`
	ContentType string        `json:"contentType,omitempty"`
	Cpuid       int
}

func main() {
	fdServer := flag.String("s", "http://someurl/blalbabla", "Url endpoint for face detection api")
	flag.Parse()
	var urlList []string
	for _, urlSource := range flag.Args() {
		urlList = append(urlList, getAListFromAUrl(strings.TrimSpace(urlSource))...)
	}
	resp := faceDetectImages(*fdServer, urlList)
	fmt.Println(resp)
	var i []Fd
	err := json.Unmarshal([]byte(resp), &i)
	if err != nil {
		fmt.Println(err)
	}
	for _, v := range i {
		fmt.Println("<h1>", v.NumFaces, "</h1>", "<img src='"+v.Url+"'>")
	}
}
