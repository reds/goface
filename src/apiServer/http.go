package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Resp struct {
	Num         int    `json:"num"`
	Url         string `json:"url"`
	ErrorNum    int    `json:"errorNum"`
	ErrorString string `json:"errorString"`
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
	faceDetect(resp.Header.Get("content-type"), body)
	return ret
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
	log.Println(string(body))
	if err == nil {
		b := bytes.NewBuffer(body)
		for {
			line, err := b.ReadString('\n')
			if err != nil {
				break
			}
			log.Println(line)
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

func main() {
	http.HandleFunc("/1/api/facedetect/url", handleUrls) // process a list of urls
	http.ListenAndServe(":8080", nil)
}
