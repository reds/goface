package main

/*
#cgo LDFLAGS: -lcv -lhighgui -lm

#include "opencv.c"
*/
import "C"

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unsafe"
)

func faceDetect(typ string, body []byte) []Rect {
	fn, _ := saveFile(typ, body)
	faces := C.process_image(C.CString(fn + "/orig"))
	list := C.GoString(faces)
	C.free(unsafe.Pointer(faces))
	var f []Rect
	for _, l := range strings.Split(list, "\n") {
		l = strings.TrimSpace(l)
		var x, y, w, h int
		n, err := fmt.Fscan(strings.NewReader(l), &x, &y, &w, &h)
		if err != nil && n != 4 {
			log.Println("error", err)
			continue
		}
		f = append(f, Rect{x, y, w, h})
	}
	return f
}

func saveFile(typ string, body []byte) (string, error) {
	h := sha1.New()
	h.Write(body)
	s := fmt.Sprintf("%x", h.Sum(nil))
	// 3 levels just for fun
	path := fmt.Sprintf("%s/%s/%s/%s", *imageRoot, s[:2], s[2:8], s[8:])
	err := os.MkdirAll(path, 0644)
	if err != nil {
		return "", err
	}
	ioutil.WriteFile(path+"/orig", body, 0600)
	return path, nil
}

func init() {
	C.init(C.CString(*haarFile))
}
