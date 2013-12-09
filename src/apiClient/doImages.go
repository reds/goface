package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/template"
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

type fdSorter struct {
	fds []*Fd
}

func (s *fdSorter) Len() int {
	return len(s.fds)
}

func (s *fdSorter) Less(a, b int) bool {
	if s.fds[a].NumFaces < 1 {
		return true
	}
	if s.fds[b].NumFaces < 1 {
		return true
	}
	return s.fds[a].Faces[0].H > s.fds[b].Faces[0].H
}

func (s *fdSorter) Swap(a, b int) {
	s.fds[a], s.fds[b] = s.fds[b], s.fds[a]
}

func main() {
	fdServer := flag.String("s", "http://someurl/blalbabla", "Url endpoint for face detection api")
	nImages := flag.Int("n", -1, "Number of images to process")
	flag.Parse()
	var urlList []string
	for _, urlSource := range flag.Args() {
		urlList = append(urlList, getAListFromAUrl(strings.TrimSpace(urlSource))...)
	}
	if *nImages > 0 && *nImages < len(urlList) {
		urlList = urlList[:*nImages]
	}

	resp := faceDetectImages(*fdServer, urlList)
	var i []*Fd
	err := json.Unmarshal([]byte(resp), &i)
	if err != nil {
		fmt.Println(err)
	}
	sort.Sort(&fdSorter{i})
	fmt.Println(topOfPage)
	tmpl := template.Must(template.New("test").Parse(canvasTmpl))

	for _, v := range i {
		if v.NumFaces == 1 {
			//			fmt.Println("<img src='"+v.Url+"'>(", v.NumFaces, ",", v.Faces[0].H, ")")
			err := tmpl.Execute(os.Stdout, v)
			if err != nil {
				panic(err)
			}
		}
	}
}

var topOfPage = `<script>
function loadImage ( n, url, x, y, w, h ) {
      var ctx = document.getElementById("myCanvas" + n).getContext('2d');
      var imageObj = new Image();

      imageObj.onload = function() {
console.log(imageObj.x, imageObj.y, imageObj.width, imageObj.height);
console.log(arguments);
        ctx.drawImage(imageObj, 0, 0);
ctx.beginPath();
ctx.lineWidth = 2;
ctx.strokeStyle = "green";
ctx.rect(x, y, w, h );
ctx.stroke();
      };
      imageObj.src = url;
}
    </script>

`

var canvasTmpl = `
<canvas id="myCanvas{{.Num}}" width="240" height="180"></canvas>
<script>loadImage ( {{.Num}}, "{{.Url}}", {{range .Faces}}{{.X}}, {{.Y}}, {{.W}}, {{.H}}{{end}} );</script>
`
