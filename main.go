package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bmaupin/go-epub"
	"io"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

var LINK = "Not Updated Yet"

// takes a url returns a goquery Document
func getDocFromURL(url string) *goquery.Document {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

//gets most recent chapter from The Wandering Inn, returns selection selection that holds url
func findNewChapSelection() *goquery.Selection {
	doc := getDocFromURL("https://wanderinginn.com/")
	return doc.Find("#latest-chapter-display")
}

//Finds link inside of a selection, returns the url 
func parseChapSelection(selection *goquery.Selection) string {
	url, ok := selection.Find("a").Attr("href")
	if !ok {
		return "404"
	}
	return url
}

//Returns a slice of all the paragraphs, one in each index. 
//Returns a string containing the title of the chapter
func getChapterBody(url string) ([]string, string) {
	//select all <p> sections
	doc := getDocFromURL(url)
	title := doc.Find(".entry-title").Text()
	sel := doc.Find(".entry-content")
	chapter := make([]string, 0)
	sel.Children().Each(func(i int, s *goquery.Selection) {
		paragraph := s.Text()
		if i < 50 {
			fmt.Println(paragraph)
			fmt.Println("++++++++++++++")
			fmt.Print(i)
		}
		chapter = append(chapter, paragraph)

	})
	return chapter, title
}

//Calls each function needed to get a book from a URL
func buildBook(chapterURL string){
	paragraphs, title := getChapterBody(chapterURL)
	title =	strings.ReplaceAll(title, ".", "-")
	strings.
	if(os.Getenv("BOOKVERSION") == title){
		log.Println("Same version of book")
		return
	}
	//Sets tile and author of ebook
	book := epub.NewEpub("TWI" + title)
	book.SetAuthor("Pirateaba")
	

	//Combine each paragraph in chapter's body[],
	var sectionStrBuilder strings.Builder //More efficient string concatenation
	for _, paragraph := range paragraphs {
		paragraph = strings.ReplaceAll(paragraph, "<", "[")
		paragraph = strings.ReplaceAll(paragraph, ">", "]")
		sectionStrBuilder.WriteString("<p>" + paragraph + "</p>") //adds each paragraph to builder
	}
	//convert builder to string, add that section to ebook.epub
	_, err := book.AddSection(sectionStrBuilder.String(), "", "", "")
	if err != nil {
		log.Println("error in adding section")
		log.Fatal(err)
	}

	//create and write .epub file
	title = strings.ReplaceAll(title, " ", "_")
	//Set environment variable 
	os.Setenv("BOOKVERSION", title)
	filePath := "assets/TWI" + title + ".epub"
	errWrite := book.Write(filePath)
	LINK = "TWI" + title + ".epub"
	if errWrite != nil {
		log.Fatal(err)
	}
}
// templateData provides template parameters.

func handlerStart(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("NAME")
	if name == "" {
		name = "World"
	}
	fmt.Fprintf(w, "BasePage %s!\n", name)
}

func handlerBuild(w http.ResponseWriter, r *http.Request) {
	selection := findNewChapSelection()
	chapterURL := parseChapSelection(selection)
	buildBook(chapterURL)	
}

func handlerGetURL(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		// Handle error
		log.Printf("Error parsing form: %s", err.Error())
		return
	}

	url := r.Form.Get("chapterURL")
	if url != "" {
		buildBook(url)
		log.Print(url)
	} else {
		log.Printf("Empty URL string")
	}
}

func handlerLinkUpdate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(LINK))
	data.LatestChapter = LINK
}

//helloRunHandler responds to requests by rendering an HTML page.
func helloRunHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.Execute(w, data); err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		log.Printf("template.Execute: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}


type templateData struct {
	Service  string
	Revision string
	LatestChapter string
}

// Variables used to generate the HTML page.
var (
	data templateData
	tmpl *template.Template
)

func main() {
	// Initialize template parameters.

	// os.Setenv("BOOKVERSION", "NotRecievedYet")
	// Prepare template for execution.
	tmpl = template.Must(template.ParseFiles("index.html"))
	data = templateData{
		Service:  "Cloud Code",
		Revision: "0.04.2",
		LatestChapter: "NotRecievedYet",
	}

	// Define HTTP server.
	http.HandleFunc("/", helloRunHandler)
	http.HandleFunc("/build", handlerBuild)
	http.HandleFunc("/getLinkUpdate", handlerLinkUpdate)
	http.HandleFunc("/getChapURL", handlerGetURL)
	
	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// PORT environment variable is provided by Cloud Run.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Print("Hello from Cloud Run! The container started successfully and is listening for HTTP requests on $PORT")
	log.Printf("Listening on port %s", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}


