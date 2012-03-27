package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"text/template"
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
	Date        string
	MachineDate string
	Keywords    string
	Description string
	Tags        []string
}

type Tag struct {
	Title string
	Posts map[string]*Post
}

type Sidebar struct {
	Recent []Post
	Tags   map[string]*Tag
	Pages  map[string]*Page
}

const assetPath = len("/")
const pagePath = len("/page/")
const tagPath = len("/tag/")
const postPath = len("/")

const maxPosts = 10 // Number posts to display on homepage

// Pages
var pages = make(map[string]*Page)
var pageTemplates = make(map[string]*template.Template)

// Posts
var posts = make(map[string]*Post)
var postsJSON []Post // Need this so that there is an ordered list of posts
var postTemplates = make(map[string]*template.Template)

// Templates
var layoutTemplates *template.Template
var errorTemplates *template.Template
var rssTemplate *template.Template
var sidebarAssets *Sidebar

// Tags
var tags = make(map[string]*Tag)

// Init Function to Load Template Files and JSON Dict to Cache
func init() {
	loadTemplates()
	loadPages()
	loadPosts()
	loadTags()

	n := 5

	if len(postsJSON) < 5 {
		n = len(postsJSON)
	}

	sidebarAssets = &Sidebar{postsJSON[0:n], tags, pages}
}

// Load The Tags Map
func loadTags() {
	for i := 0; i < len(postsJSON); i++ {

		for t := 0; t < len(postsJSON[i].Tags); t++ {

			_, ok := tags[postsJSON[i].Tags[t]]
			if ok {
				tags[postsJSON[i].Tags[t]].Posts[postsJSON[i].Title] = &postsJSON[i]
			} else {
				tagPosts := make(map[string]*Post)
				tagPosts[postsJSON[i].Title] = &postsJSON[i]
				tags[postsJSON[i].Tags[t]] = &Tag{postsJSON[i].Tags[t], tagPosts}
			}

		}
	}
}

// Load Pages Dict and Templates
func loadPages() {
	pagesRaw, _ := ioutil.ReadFile("data/pages.json")
	var pagesJSON []Page
	err := json.Unmarshal(pagesRaw, &pagesJSON)
	if err != nil {
		panic("Could not parse Pages JSON!")
	}

	for i := 0; i < len(pagesJSON); i++ {
		pages[pagesJSON[i].Slug] = &pagesJSON[i]
	}

	for _, tmpl := range pages {
		t := template.Must(template.ParseFiles("./pages/" + tmpl.Slug + ".html"))
		pageTemplates[tmpl.Slug] = t
	}
}

// Load Posts Dict and Templates
func loadPosts() {
	postsRaw, _ := ioutil.ReadFile("data/posts.json")
	err := json.Unmarshal(postsRaw, &postsJSON)
	if err != nil {
		panic("Could not parse Posts JSON!")
	}

	for i := 0; i < len(postsJSON); i++ {
		posts[postsJSON[i].Slug] = &postsJSON[i]
	}

	for _, tmpl := range posts {
		t := template.Must(template.ParseFiles("./posts/" + tmpl.Slug + ".html"))
		postTemplates[tmpl.Slug] = t
	}
}

// Load Layout and Error Templates
func loadTemplates() {
	layoutTemplates = template.Must(template.ParseFiles("./templates/layouts.html"))
	errorTemplates = template.Must(template.ParseFiles("./errors/404.html", "./errors/505.html"))
	rssTemplate = template.Must(template.ParseFiles("./templates/rss.xml"))
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
		errorTemplates.ExecuteTemplate(w, "404", nil)
		return
	}

	// Find the page
	p := pages[slug]

	// Header
	layoutTemplates.ExecuteTemplate(w, "Header", p)

	// Sidebar
	layoutTemplates.ExecuteTemplate(w, "Sidebar", sidebarAssets)

	// Page Template
	err := pageTemplates[slug].Execute(w, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorTemplates.ExecuteTemplate(w, "505", nil)
		return
	}

	// Footer
	layoutTemplates.ExecuteTemplate(w, "Footer", nil)
}

// Post Handler
func postHandler(w http.ResponseWriter, r *http.Request) {
	// Get the post slug, use 'index' if no slug is present
	slug := r.URL.Path[postPath:]
	if slug == "" {
		indexHandler(w, r)
		return
	}

	// Check that the post exists and return 404 if it doesn't
	_, ok := posts[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		errorTemplates.ExecuteTemplate(w, "404", nil)
		return
	}

	// Find the post
	p := posts[slug]

	// Header
	layoutTemplates.ExecuteTemplate(w, "Header", p)

	// Sidebar
	layoutTemplates.ExecuteTemplate(w, "Sidebar", sidebarAssets)

	// Post Template
	err := postTemplates[slug].Execute(w, p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorTemplates.ExecuteTemplate(w, "505", nil)
		return
	}

	// Comments
	layoutTemplates.ExecuteTemplate(w, "Comments", nil)

	// Footer
	layoutTemplates.ExecuteTemplate(w, "Footer", nil)
}

// Asset Handler Serves CSS, JS and Images
func assetHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[assetPath:])
}

func archiveHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{"archive", "Archive", "", ""}

	// Header
	layoutTemplates.ExecuteTemplate(w, "Header", p)

	// Sidebar
	layoutTemplates.ExecuteTemplate(w, "Sidebar", sidebarAssets)

	// Archives
	layoutTemplates.ExecuteTemplate(w, "Archive", postsJSON)

	// Footer
	layoutTemplates.ExecuteTemplate(w, "Footer", p)
}

func tagHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[tagPath:]
	if slug == "" {
		indexHandler(w, r)
		return
	}

	// Check that the tag exists and return 404 if it doesn't
	_, ok := tags[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		errorTemplates.ExecuteTemplate(w, "404", nil)
		return
	}

	p := &Page{"/tag/" + slug, "Posts Tagged #" + slug, "", ""}

	// Header
	layoutTemplates.ExecuteTemplate(w, "Header", p)

	// Sidebar
	layoutTemplates.ExecuteTemplate(w, "Sidebar", sidebarAssets)

	for _, tmpl := range tags[slug].Posts {
		postTemplates[tmpl.Slug].Execute(w, posts[tmpl.Slug])
	}

	// Footer
	layoutTemplates.ExecuteTemplate(w, "Footer", p)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	p := pages["index"]

	// Header
	layoutTemplates.ExecuteTemplate(w, "Header", p)

	// Sidebar
	layoutTemplates.ExecuteTemplate(w, "Sidebar", sidebarAssets)

	// Show Recent Posts
	for i, tmpl := range postsJSON {
		if i >= maxPosts {
			break
		}
		postTemplates[tmpl.Slug].Execute(w, posts[tmpl.Slug])
	}

	// Footer
	layoutTemplates.ExecuteTemplate(w, "Footer", p)
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml")
	rssTemplate.Execute(w, postsJSON)
}

// Starts Server and Routes Requests
func main() {
	http.HandleFunc("/archive", archiveHandler)
	http.HandleFunc("/page/", pageHandler)
	http.HandleFunc("/tag/", tagHandler)
	http.HandleFunc("/assets/", assetHandler)
	http.HandleFunc("/rss", rssHandler)
	http.HandleFunc("/", postHandler)
	http.ListenAndServe(":9981", nil)
}
