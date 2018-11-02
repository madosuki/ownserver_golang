package pkg

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"text/template"
)

type serverMethods interface {
	not_found(w http.ResponseWriter, req *http.Request)
	sned_image(w http.ResponseWriter, req *http.Request, str string)
	handler(w http.ResponseWriter, req *http.Request)
}

type server struct {
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
		s.write_log("Template Parse Error. from function not_found(w http.ResponseWriter, req *http.Request)")
		return
	}

	w.Header().Set("Content-Type:", "text/html")

	err = tmp.Execute(w, nil)

	if err != nil {
		s.write_log("Template Execute Error.")
	}
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

type send_byte_struct struct {
	filename string
	is_css   bool
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

	w.Header().Set("Content-Type", mime)

	w.Write(tmp)
}

func (s *server) Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		// w.Header().Set("Content-Type", "text/html; charset=utf-8")

		reg := regexp.MustCompile(`[a-z0-9]*.[jpg | png | gif]`)
		css := regexp.MustCompile(`css/[a-z0-9]*.css`)
		str := req.URL.Path

		if reg.MatchString(str) {
			data := send_byte_struct{str[1:], false}
			s.send_byte(w, req, data)
		} else if css.MatchString(str) {
			data := send_byte_struct{str[1:], true}
			s.send_byte(w, req, data)
		} else {

			page := page{"Alice in Wonderland"}
			tmp := template.Must(template.ParseFiles(
				"base.html",
				"index.html"))

			// tmp, err := template.New("new").Parse("<h1>{{.Title}}</h1><img src='test.jpg'>")

			/*
				if err != nil {
					s.write_log("Template Parse Error. from function Handler")
				}
			*/

			w.Header().Set("Content-Type", "text/html")

			err := tmp.Execute(w, page)

			if err != nil {
				s.write_log("Template Execute Error. from function Handler")
			}
		}

	}
}

var instance *server = new(server)

func GetInstance() *server {
	return instance
}