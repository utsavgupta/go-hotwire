package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type Person struct {
	Name string `form:"name"`
}

func GenerateTemplates() *template.Template {
	t := template.Must(template.ParseGlob("templates/layout/*.gohtml"))
	t = template.Must(t.ParseGlob("templates/partials/*.gohtml"))
	t = template.Must(t.ParseGlob("templates/*.gohtml"))

	return t
}

func NewRouter() *mux.Router {
	return mux.NewRouter()
}

func PrepareRoutesWithTemplates(router *mux.Router, t *template.Template) {
	router.Handle("/", newIndex(t)).Methods(http.MethodGet)
	router.Handle("/greet", newGreetings(t)).Methods(http.MethodPost)
}

func newIndex(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeHeaders(w, http.StatusOK)
		err := templates.ExecuteTemplate(w, "index.gohtml", nil)

		if err != nil {
			fmt.Println(err)
		}
	}
}

func newGreetings(templates *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if e := r.ParseForm(); e != nil {
			fmt.Fprintf(os.Stderr, "could not parse form")
			writeHeaders(w, http.StatusBadRequest)
			return
		}

		name := r.FormValue("name")

		time.Sleep(100 * time.Millisecond)

		writeHeaders(w, http.StatusOK)
		err := templates.ExecuteTemplate(w, "greetings_partial", Person{name})

		if err != nil {
			fmt.Println(err)
		}
	}
}

func writeHeaders(w http.ResponseWriter, status int) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(status)
}

func main() {
	router := NewRouter()
	templates := GenerateTemplates()
	PrepareRoutesWithTemplates(router, templates)
	http.ListenAndServe(":8080", router)
}
