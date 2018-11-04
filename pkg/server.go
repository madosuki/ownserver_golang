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

type send_byte_struct struct {
	filename string
	is_css   bool
}

type serverMethods interface {
	write_log(str string)
	not_found(w http.ResponseWriter, req *http.Request)
	get_filesize(name string) int64
	read_file(name string, size int64) []byte
	encode_byte_to_gzip(buf []byte) (*bytes.Buffer, bool)
	last_send_process(w http.ResponseWriter, req *http.Request, mime string, buf []byte)
	sned_byte(w http.ResponseWriter, req *http.Request, data send_byte_struct)
	send_html(w http.ResponseWriter, req *http.Request)
	Handler(w http.ResponseWriter, req *http.Request)
}

type server struct {
	is_index   *regexp.Regexp
	is_picture *regexp.Regexp
	is_css     *regexp.Regexp
}

type page struct {
	Title string
}

func (s *server) write_log(str string) {
	file, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Print(err)
	}

	defer file.Close()

	fmt.Fprintf(file, str+"\n")
}

func (s *server) not_found(w http.ResponseWriter, req *http.Request) {
	tmp, err := template.ParseFiles("404.html")

	if err != nil {
		s.write_log("Template Parse Error, from function not_found(w http.ResponseWriter, req *http.Request)")
		return
	}

	buf := new(bytes.Buffer)
	err = tmp.Execute(buf, nil)

	if err != nil {
		s.write_log("Template Execute Error.")
	}

	s.last_send_process(w, req, "text/html", buf.Bytes())
}

func (s *server) get_filesize(name string) int64 {
	exist, err := os.Stat(name)

	if err != nil {
		return 0
	}

	return exist.Size()
}

func (s *server) read_file(name string, size int64) []byte {
	file, err := os.Open(name)
	defer file.Close()
	if err != nil {
		s.write_log("File Open Error. from function read_file(name sting, size int64) []byte")
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

func (s *server) encode_byte_to_gzip(buf []byte) (*bytes.Buffer, bool) {
	tmp := new(bytes.Buffer)
	gw := gzip.NewWriter(tmp)

	_, err := gw.Write(buf)

	if err != nil {
		s.write_log("Error, gzip encode execute failed.")
		return tmp, false
	}

	err = gw.Close()

	if err != nil {
		s.write_log("Error, gzip.NewWriter().Close() failed.")
		return tmp, false
	}

	return tmp, true
}

func (s *server) last_send_process(w http.ResponseWriter, req *http.Request, mime string, buf []byte) {

	if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
		tmp, check := s.encode_byte_to_gzip(buf)
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

func (s *server) send_byte(w http.ResponseWriter, req *http.Request, data send_byte_struct) {
	size := s.get_filesize(data.filename)

	if size == 0 {
		s.not_found(w, req)
		return
	}

	tmp := s.read_file(data.filename, size)

	mime := http.DetectContentType(tmp)

	if data.is_css {
		mime = "text/css"
	}

	fmt.Println(mime)

	s.last_send_process(w, req, mime, tmp)
}

func (s *server) send_html(w http.ResponseWriter, req *http.Request) {

	url := req.URL.Path

	if s.is_index.MatchString(url) {

		page := page{"Alice in Wonderland"}

		tmp := template.Must(template.ParseFiles(
			"base.tmpl",
			"index.tmpl"))

		w.Header().Set("Content-Type", "text/html")

		buf := new(bytes.Buffer)
		err := tmp.Execute(buf, page)

		if err != nil {
			s.write_log("Template Execute Error, from function Handler")
		}

		s.last_send_process(w, req, "text/html", buf.Bytes())

	} else {
		s.not_found(w, req)
	}
}

func (s *server) Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {

		path := req.URL.Path

		fmt.Println(path)

		if s.is_picture.MatchString(path) {
			data := send_byte_struct{path[1:], false}
			s.send_byte(w, req, data)
		} else if s.is_css.MatchString(path) {
			data := send_byte_struct{path[1:], true}
			s.send_byte(w, req, data)
		} else {
			s.send_html(w, req)
		}

	}
}

var instance *server = constructor()

func constructor() *server {
	i := new(server)

	i.is_index = regexp.MustCompile(`/|index.html`)
	i.is_picture = regexp.MustCompile(`[a-z0-9]*.jpg|png|gif|webp`)
	i.is_css = regexp.MustCompile(`css/[a-z0-9]*.css`)

	return i
}

func GetInstance() *server {
	return instance
}
