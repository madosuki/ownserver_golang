package pkg

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"
)

type sendByteStruct struct {
	filename string
	isCSS    bool
}

type serverMethods interface {
	writeLog(str string)
	notFound(w http.ResponseWriter, req *http.Request)
	getFileSize(name string) int64
	readFile(name string, size int64) []byte
	encodeByteToGzip(buf []byte) (*bytes.Buffer, bool)
	lastSendProcess(w http.ResponseWriter, req *http.Request, mime string, buf []byte)
	snedByte(w http.ResponseWriter, req *http.Request, data sendByteStruct)
	sendHtml(w http.ResponseWriter, req *http.Request)
	Handler(w http.ResponseWriter, req *http.Request)
}

type server struct {
	isIndex   *regexp.Regexp
	isPicture *regexp.Regexp
	isCSS     *regexp.Regexp
	isVideo   *regexp.Regexp
}

type page struct {
	Title string
}

func (s *server) writeLog(str string) {
	file, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Print(err)
	}

	defer file.Close()

	fmt.Fprintf(file, str+"\n")
}

func (s *server) notFound(w http.ResponseWriter, req *http.Request) {
	tmp, err := template.ParseFiles("404.html")

	if err != nil {
		s.writeLog("Template Parse Error, from function notFound(w http.ResponseWriter, req *http.Request)")
		return
	}

	buf := new(bytes.Buffer)
	err = tmp.Execute(buf, nil)

	if err != nil {
		s.writeLog("Template Execute Error.")
	}

	s.lastSendProcess(w, req, "text/html", buf.Bytes())
}

func (s *server) getFileSize(name string) int64 {
	exist, err := os.Stat(name)

	if err != nil {
		return 0
	}

	return exist.Size()
}

func (s *server) readFile(name string, size int64) []byte {
	file, err := os.Open(name)
	defer file.Close()
	if err != nil {
		s.writeLog("File Open Error. from function readFile(name sting, size int64) []byte")
	}

	buf := make([]byte, size)
	for {
		n, err := file.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			break
		}
	}

	return buf
}

func (s *server) encodeByteToGzip(buf []byte) (*bytes.Buffer, bool) {
	tmp := new(bytes.Buffer)
	gw := gzip.NewWriter(tmp)

	_, err := gw.Write(buf)

	if err != nil {
		s.writeLog("Error, gzip encode execute failed.")
		return tmp, false
	}

	err = gw.Close()

	if err != nil {
		s.writeLog("Error, gzip.NewWriter().Close() failed.")
		return tmp, false
	}

	return tmp, true
}

func (s *server) lastSendProcess(w http.ResponseWriter, req *http.Request, mime string, buf []byte) {

	if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		tmp, check := s.encodeByteToGzip(buf)
		if check {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Type", mime)
			tmp.WriteTo(w)
		}
	} else {
		w.Header().Set("Content-Type", mime)
		w.Write(buf)
	}
}

func (s *server) sendByte(w http.ResponseWriter, req *http.Request, data sendByteStruct) {
	size := s.getFileSize(data.filename)

	if size == 0 {
		s.notFound(w, req)
		return
	}

	tmp := s.readFile(data.filename, size)

	mime := http.DetectContentType(tmp)

	if data.isCSS {
		mime = "text/css"
	}

	fmt.Println(mime)

	s.lastSendProcess(w, req, mime, tmp)
}

func (s *server) sendHtml(w http.ResponseWriter, req *http.Request) {

	url := req.URL.Path

	if s.isIndex.MatchString(url) {

		page := page{"Alice in Wonderland"}

		tmp := template.Must(template.ParseFiles(
			"base.tmpl",
			"index.tmpl"))

		w.Header().Set("Content-Type", "text/html")

		buf := new(bytes.Buffer)
		err := tmp.Execute(buf, page)

		if err != nil {
			s.writeLog("Template Execute Error, from function Handler")
		}

		s.lastSendProcess(w, req, "text/html", buf.Bytes())

	} else if url == "/movie" {
		tmp := template.Must(template.ParseFiles(
			"base.tmpl",
			"movie.tmpl"))
		w.Header().Set("Content-Type", "text/html")

		buf := new(bytes.Buffer)
		err := tmp.Execute(buf, nil)

		if err != nil {
			s.writeLog("Template execute error.")
		}

		s.lastSendProcess(w, req, "text/html", buf.Bytes())

	} else {
		s.notFound(w, req)
	}
}

func (s *server) Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {

		path := req.URL.Path

		fmt.Println(path)

		if s.isPicture.MatchString(path) {
			data := sendByteStruct{path[1:], false}
			s.sendByte(w, req, data)
		} else if s.isCSS.MatchString(path) {
			data := sendByteStruct{path[1:], true}
			s.sendByte(w, req, data)
		} else if s.isVideo.MatchString(path) {
			data := sendByteStruct{path[1:], false}
			s.sendByte(w, req, data)
		} else {
			s.sendHtml(w, req)
		}

	}
}

var instance *server = constructor()

func constructor() *server {
	i := new(server)

	i.isIndex = regexp.MustCompile(`index.html$|^/$`)
	i.isPicture = regexp.MustCompile(`/[[:word:]]*.jpg|png|gif|webp$`)
	i.isCSS = regexp.MustCompile(`/css/[[:word:]]*.css$`)
	i.isVideo = regexp.MustCompile(`videos/[[:word:]]*.mp4|m2ts|webm|mpd|m4s$`)

	return i
}

func GetInstance() *server {
	return instance
}
