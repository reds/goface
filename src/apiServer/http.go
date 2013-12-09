package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type Resp struct {
	Num         int           `json:"num"`
	Url         string        `json:"url"`
	ErrorNum    int           `json:"errorNum"`
	ErrorString string        `json:"errorString,omitempty"`
	NumFaces    int           `json:"numFaces,omitempty"`
	Faces       []Rect        `json:"faces,omitempty"`
	Time        time.Duration `json:"time,omitempty"`
	ContentType string        `json:"contentType,omitempty"`
	Cpuid       int
	data        []byte
	resp        chan *Resp
}

var fdChan = make(chan *Resp)

func doOneUrlLoop(id int) {
	for ret := range fdChan {
		typ := "jpg" // should look at ret.ContentType
		path, err := saveFile(typ, ret.data)
		if err != nil {
			log.Println(err)
			continue
		}
		ret.Url = "/static/images" + path[len(*imageRoot):] + "/orig." + typ
		ret.Cpuid = id
		ret.ErrorNum = 0
		ret.ErrorString = ""
		t1 := time.Now()
		ret.Faces = faceDetect(path + "/orig." + typ)
		ret.NumFaces = len(ret.Faces)
		ret.Time = time.Since(t1)
		ret.resp <- ret
	}
}

func doOneUrl(n int, url string) *Resp {
	ret := &Resp{Num: n, Url: url, ErrorNum: 8, ErrorString: "Url Not Found"}
	resp, err := http.Get(url)
	if err != nil {
		ret.ErrorNum = 7
		ret.ErrorString = err.Error()
		return ret
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ret.ErrorNum = 6
		ret.ErrorString = err.Error()
		return ret
	}
	respChan := make(chan *Resp)
	fdChan <- &Resp{Num: n, Url: url, ContentType: resp.Header.Get("content-type"), data: body, resp: respChan}
	return <-respChan
}

func getUrlList(r *http.Request) []string {
	r.ParseForm()
	a := []string{}
	for k, v := range r.Form {
		if strings.Index(k, "url") != 0 {
			continue
		}
		vv := ""
		if len(v) > 0 {
			vv = v[0]
		}
		a = append(a, vv)
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	if err == nil {
		b := bytes.NewBuffer(body)
		for {
			line, err := b.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimSpace(line)
			a = append(a, line)
		}
	}
	return a
}

func handleUrls(w http.ResponseWriter, r *http.Request) {
	res := []*Resp{}
	for i, v := range getUrlList(r) {
		res = append(res, doOneUrl(i, v))
	}
	buf, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
	}
	w.Write(buf)
}

func static(w http.ResponseWriter, r *http.Request) {
	log.Println(r)
}

var (
	imageRoot = flag.String("i", "/tmp/facedetect/static/images", "Root directory for processed images")
	haarFile  = flag.String("h", "haarcascade_frontalface_default.xml", "File containing the trained haar parameters used by OpenCV")
)

func main() {
	flag.Parse()
	for i := 0; i < 10; i++ {
		go doOneUrlLoop(i)
	}
	http.Handle("/static/", http.FileServer(http.Dir("/tmp/facedetect/")))
	http.HandleFunc("/1/api/facedetect/url", handleUrls) // process a list of urls
	panic(http.ListenAndServe(":8088", nil))
}
