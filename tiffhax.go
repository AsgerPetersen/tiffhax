package main

import (
	"flag"
	"github.com/emilyselwood/tiffhax/parser/tiff"
	"github.com/emilyselwood/tiffhax/payload"
	"github.com/skratchdot/open-golang/open"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
)

func parseFile(filePath string) payload.Payload {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Could not open file: %s", err)
	}
	defer f.Close()

	sections, err := tiff.ParseFile(f)
	if err != nil {
		log.Fatalf("Could not open parse: %s", err)
	}

	return payload.Payload{
		Title: "tiff hax", FileName: filePath, Sections: sections,
	}
}

func setupHttpServer(data payload.Payload, debug bool) net.Listener {
	var templates *template.Template
	var err error
	if !debug {
		templates, err = template.ParseFiles(
			"templates/index.template.html",
			"templates/header.template.html",
		)
		if err != nil {
			log.Fatalf("Could not parse template files %s", err)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simply write some test data for now
		if debug {
			templates, err = template.ParseFiles(
				"templates/index.template.html",
				"templates/header.template.html",
			)
			if err != nil {
				log.Fatalf("Could not parse template files %s", err)
			}
		}

		if err := templates.Execute(w, data); err != nil {
			log.Printf("Error writing template: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// run the webserver
	l, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}

	return l
}

func main() {
	// set up, get flags etc
	debug := flag.Bool("debug", false, "reload the templates on every request.")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.PrintDefaults()
		log.Fatal("a filename is required")
	}

	// open the file and parse it to create the payload information.
	data := parseFile(flag.Arg(0))

	// setup the http server
	l := setupHttpServer(data, *debug)

	// The browser can connect now because the listening socket is open.
	err := open.Start("http://localhost:3000/")
	if err != nil {
		log.Println(err)
	}

	// Start the blocking server loop.
	log.Fatal(http.Serve(l, http.DefaultServeMux))
}