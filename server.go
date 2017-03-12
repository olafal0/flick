package flick

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/kabukky/httpscerts"
)

// Context is an extended version of ReponseWriter
type Context struct {
	Wr      http.ResponseWriter
	Req     *http.Request
	Queries map[string][]string
}

func (c *Context) Write(data []byte) {
	reader := bytes.NewReader(data)
	name := c.Req.RequestURI
	http.ServeContent(c.Wr, c.Req, name, time.Now(), reader)
}

// Serve starts the webserver
func Serve(addr string) {
	PrepareStatics()
	log.Printf("Serving on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// ServeTLS uses a supplied certfile and keyfile to serve HTTPS
func ServeTLS(addr, certfile, keyfile string) {
	PrepareStatics()
	log.Printf("Serving on %s\n", addr)
	log.Fatal(http.ListenAndServeTLS(addr, certfile, keyfile, nil))
}

// ServeTLSSelfSign auto-generates a self-signed certificate. For testing purposes only.
func ServeTLSSelfSign(addr string) {
	err := httpscerts.Check("cert.pem", "key.pem")
	//If they are not available, generate new ones.
	if err != nil {
		err := httpscerts.Generate("cert.pem", "key.pem", addr)
		if err != nil {
			log.Fatal("Error: Couldn't create https certs.")
		}
	}

	PrepareStatics()
	log.Printf("Serving on %s\n", addr)
	log.Fatal(http.ListenAndServeTLS(addr, "cert.pem", "key.pem", nil))
}

// Get takes a pattern string and a function(*http.Request)
// and adds it to the DefaultServeMux
func Get(pattern string, handler func(c *Context)) {

	http.HandleFunc(pattern,
		func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			// make sure this is actually supposed to be a GET request
			if r.Method != "" && r.Method != "GET" {
				// use reflection to get the name of the handler method
				methodName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
				log.Printf("Warning: GET handler for function %s got non-GET method type", methodName)
			}
			handler(&Context{w, r, r.URL.Query()})
			elapsed := time.Since(start)
			log.Printf("%s %s: %s", r.Proto, pattern, elapsed)
		})

}

func PrepareStatics() {
	files, err := ioutil.ReadDir("./static/")
	if err != nil {
		log.Print(err)
	}
	fmt.Print("Adding static files:\n")
	for _, f := range files {
		fmt.Println(f.Name())
		serveStaticFile(f.Name(), f.ModTime())
	}
}

func serveStaticFile(filename string, modtime time.Time) {
	path := "static/" + filename
	// get contents of file
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Error serving %s: %v", path, err)
		return
	}
	contentsReader := bytes.NewReader(contents)
	Get("/"+filename,
		func(c *Context) {
			http.ServeContent(c.Wr, c.Req, filename, modtime, contentsReader)
		})
}
