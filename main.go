package main

import (
	"http"
	"template"
	"io/ioutil"
	"json"
)

type Page struct {
	Slug        string
	Title       string
	Keywords    string
	Description string
}

type Post struct {
	Title       string
	Slug        string
	Date				string
	Keywords    string
	Description string
	Tags				[]string
}

const assetPath = len("/")
const pagePath = len("/page/")
const postPath = len("/")

// Pages
var pages = make(map[string]*Page)
var pageTemplates = make(map[string]*template.Template)

// Posts
var posts = make(map[string]*Post)
var postTemplates = make(map[string]*template.Template)

// Templates
var layoutTemplates *template.Set
var errorTemplates *template.Set

// Init Function to Load Template Files and JSON Dict to Cache
func init() {
	// Parse Page JSON Dict
	pagesRaw, _ := ioutil.ReadFile("pages/pages.json")
	var pagesJSON []Page
	err := json.Unmarshal(pagesRaw, &pagesJSON)
	if err != nil {
		panic("Could not parse Pages JSON!")
	}

	// Put Pages into pages map
	for i := 0; i < len(pagesJSON); i++ {
		pages[pagesJSON[i].Slug] = &pagesJSON[i]
	}


	// Parse Posts JSON Dict
	postsRaw, _ := ioutil.ReadFile("posts/posts.json")
	var postsJSON []Post
	err = json.Unmarshal(postsRaw, &postsJSON)
	if err != nil {
		panic("Could not parse Posts JSON!")
	}

	// Put Posts into pages map
	for i := 0; i < len(postsJSON); i++ {
		posts[postsJSON[i].Slug] = &postsJSON[i]
	}


	// Parse and Cache Page Templates
	for _, tmpl := range pages {
		t := template.Must(template.ParseFile("./pages/" + tmpl.Slug + ".html"))
		pageTemplates[tmpl.Slug] = t
	}

	// Parse and Cache Posts Templates
	for _, tmpl := range posts {
		t := template.Must(template.ParseFile("./posts/" + tmpl.Slug + ".html"))
		postTemplates[tmpl.Slug] = t
	}

	// Parse and Cache Layout Templates
	layoutTemplates = template.SetMust(template.ParseSetFiles("templates.html"))

	// Parse and Cache Error Templates
	errorTemplates = template.SetMust(template.ParseSetFiles("./errors/404.html", "./errors/505.html"))
}

// Page Handler Constructs and Serves Pages
func pageHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page slug, use 'index' if no slug is present
	slug := r.URL.Path[pagePath:]
	if slug == "" {
		indexHandler(w, r)
		return
	}

	// Check that the page exists and return 404 if it doesn't
	_, ok := pages[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		errorTemplates.Execute(w, "404", nil)
		return
	}

	// Find the page
	p := pages[slug]

	// Header
	layoutTemplates.Execute(w, "Header", p)

	// Page Template
	err := pageTemplates[slug].Execute(w, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorTemplates.Execute(w, "505", nil)
		return
	}

	// Footer
	layoutTemplates.Execute(w, "Footer", nil)
}

// Post Handler
func postHandler(w http.ResponseWriter, r *http.Request) {
	// Get the post slug, use 'index' if no slug is present
	slug := r.URL.Path[postPath:]
	if slug == "" {
		indexHandler(w, r)
		return
	}

	// Check that the page exists and return 404 if it doesn't
	_, ok := posts[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		errorTemplates.Execute(w, "404", nil)
		return
	}

	// Find the page
	p := posts[slug]

	// Header
	layoutTemplates.Execute(w, "Header", p)

	// Post Template
	err := postTemplates[slug].Execute(w, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorTemplates.Execute(w, "505", nil)
		return
	}

	// Footer
	layoutTemplates.Execute(w, "Footer", nil)
}

// Asset Handler Serves CSS, JS and Images
func assetHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[assetPath:])
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	p := pages["index"]

	// Header
	layoutTemplates.Execute(w, "Header", p)

	// TODO Loop Each Post

	// Footer
	layoutTemplates.Execute(w, "Footer", p)
}

// Starts Server and Routes Requests
func main() {
	http.HandleFunc("/page/", pageHandler)
	http.HandleFunc("/assets/", assetHandler)
	http.HandleFunc("/", postHandler)
	http.ListenAndServe(":9981", nil)
}
