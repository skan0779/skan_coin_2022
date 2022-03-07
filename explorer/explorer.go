// Package explorer provides html functions for skancoin
package explorer

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/skan0779/skan_coin_2022/blockchain"
)

var templates *template.Template
var port string

type homepageData struct {
	Context  string
	PageName string
	Blocks   []*blockchain.Block
}

func homepage(rw http.ResponseWriter, r *http.Request) {
	data := homepageData{
		PageName: "Home",
		Context:  "Welcome to skan coin!",
		Blocks:   blockchain.Blocks(blockchain.Blockchain()),
	}
	templates.ExecuteTemplate(rw, "home", data)
}

func addpage(rw http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		templates.ExecuteTemplate(rw, "add", "")
	case "POST":
		blockchain.Blockchain().AddBlock()
		http.Redirect(rw, r, "/", http.StatusPermanentRedirect)
	}
}

// start explorer
func Start(p int) {
	port = fmt.Sprintf(":%d", p)
	handler := http.NewServeMux()

	// loading template
	templates = template.Must(template.ParseGlob("explorer/templates/pages/*.gohtml"))
	templates = template.Must(templates.ParseGlob("explorer/templates/components/*.gohtml"))

	// handler function
	handler.HandleFunc("/", homepage)
	handler.HandleFunc("/add", addpage)

	// create server-side rendering website
	fmt.Println("Start Explorer server" + port)
	log.Fatal(http.ListenAndServe(port, handler))
}
