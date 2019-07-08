package main

import (
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"gitlab.com/golang-commonmark/markdown"
)

type Page struct {
	Title string
	Body  []byte
}

type PageView struct {
	Title string
	Body  template.HTML
}

var templates = template.Must(template.ParseFiles("views/edit.html", "views/view.html", "views/layout/nav.html", "views/layout/modals.html"))
var validPath = regexp.MustCompile("^/(edit|save|view|delete)/([a-zA-Z0-9]+)$")
var dataRoot = "data"

func (p *Page) save() error {
	filename := buildPath(p.Title + ".txt")
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func buildPath(filename string) string {
	return dataRoot + "/" + filename
}

func loadPage(title string) (*Page, error) {
	filename := buildPath(title + ".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", errors.New("Invalid Page Title")
	}
	return m[2], nil // The title is the second subexpression.
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	md := markdown.New(markdown.XHTMLOutput(true))
	parsedString := md.RenderToString([]byte(p.Body))
	px := bluemonday.UGCPolicy()
	px.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	p.Body = px.SanitizeBytes([]byte(parsedString))
	pg := &PageView{Title: p.Title, Body: template.HTML(p.Body)}
	renderTemplateC(w, "view", pg)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, title string) {
	if strings.ToLower(title) == "frontpage" {
		// you can't do this;
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	pathTo := buildPath(title + ".txt")
	if _, err := os.Stat(pathTo); err == nil {
		// path/to/whatever exists
		// time to delete;
		_ = os.Remove(pathTo)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	} else if os.IsNotExist(err) {
		// path/to/whatever does *not* exist
		// just redirect to main page.
		http.Redirect(w, r, "/", http.StatusFound)
		return
	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderTemplateC(w http.ResponseWriter, tmpl string, p *PageView) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
			return
		}
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.HandleFunc("/", makeHandler(viewHandler))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/delete/", makeHandler(deleteHandler))

	log.Fatal(http.ListenAndServe(":1789", nil))
}
