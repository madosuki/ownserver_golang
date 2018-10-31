package pkg

import (
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

type Server struct {
}

type page struct {
	Title string
}

func (s *Server) not_found(w http.ResponseWriter, req *http.Request) {
	tmp, err := template.ParseFiles("404.html")

	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type:", "text/html")

	err = tmp.Execute(w, nil)

	if err != nil {
		panic(err)
	}
}

func (s *Server) send_image(w http.ResponseWriter, req *http.Request, str string) {
	exist, err_exist := os.Stat(str)

	if err_exist != nil {
		s.not_found(w, req)
		return
	}

	size := exist.Size()

	file, err := os.Open(str)
	defer file.Close()
	if err != nil {
		panic(err)
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

	w.Write(buf)
}

func (s *Server) Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		// w.Header().Set("Content-Type", "text/html; charset=utf-8")

		reg := regexp.MustCompile(`[a-z0-9]*.[jpg | png | gif]`)
		str := req.URL.Path

		if reg.MatchString(str) == true {
			s.send_image(w, req, str[1:])
		} else {

			page := page{"Alice in Wonderland"}
			// tmp, err := template.ParseFiles("index.html")

			tmp, err := template.New("new").Parse("<h1>{{.Title}}</h1><img src='test.jpg'>")

			if err != nil {
				panic(err)
			}

			err = tmp.Execute(w, page)

			if err != nil {
				panic(err)
			}
		}

	}
}
