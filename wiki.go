package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/joho/godotenv"
)

// Page represents a page in the Wiki
type Page struct {
	Prefix string
	Title  string
	Body   []byte
}

const (
	templatePath = "templates"
	pagesPath    = "pages"
)

func (p *Page) save() error {
	filename := pagesPath + "/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "pages/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Prefix: prefix, Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, prefix+"/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Prefix: prefix, Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Prefix: prefix, Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, prefix+"/view/"+title, http.StatusFound)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, r.URL.Path, http.StatusInternalServerError)
}

var templates = template.Must(template.ParseFiles(templatePath+"/edit.html", templatePath+"/view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.Error(w, r.URL.Path, http.StatusNotFound)
			return
		}
		fn(w, r, m[2])
	}
}

var prefix string
var port string
var validPath *regexp.Regexp

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	prefix = os.Getenv("APPLICATION_PREFIX")
	port = os.Getenv("ASPNETCORE_PORT")
	validPath = regexp.MustCompile("^" + prefix + "/(edit|save|view)/([a-zA-Z0-9]+)$")

	newpath := filepath.Join(".", pagesPath)
	os.MkdirAll(newpath, os.ModePerm)

	http.HandleFunc(prefix+"/view/", makeHandler(viewHandler))
	http.HandleFunc(prefix+"/edit/", makeHandler(editHandler))
	http.HandleFunc(prefix+"/save/", makeHandler(saveHandler))
	http.HandleFunc("/", notFoundHandler)

	fmt.Println("listening for a connection at http://localhost:" + port + prefix + "/view/test")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
