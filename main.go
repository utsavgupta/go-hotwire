package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/utsavgupta/go-hotwire/config"
	"github.com/utsavgupta/go-hotwire/services"
)

type Message struct {
	User string
	Bot  string
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

	cfg := config.LoadConfig()
	gemini := services.NewGeminiService(cfg)

	router.Handle("/", newIndex(t)).Methods(http.MethodGet)
	router.Handle("/chat", newChat(t, gemini)).Methods(http.MethodPost)
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

func newChat(templates *template.Template, gemini *services.GeminiService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if e := r.ParseForm(); e != nil {
			fmt.Fprintf(os.Stderr, "could not parse form")
			writeHeaders(w, http.StatusBadRequest)
			return
		}

		msg := r.FormValue("message")
		bot := "Cannot talk to Gemini at the moment. 🥲"

		geminiResp, err := gemini.GenerateResponse(msg)

		if err == nil {
			bot = geminiResp.Text
		}

		writeTurboStreamHeaders(w)
		err = templates.ExecuteTemplate(w, "_message_segment", Message{User: msg, Bot: bot})

		if err != nil {
			fmt.Println(err)
		}
	}
}

func writeHeaders(w http.ResponseWriter, status int) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(status)
}

func writeTurboStreamHeaders(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/vnd.turbo-stream.html")
	w.WriteHeader(http.StatusOK)
}

func main() {
	router := NewRouter()
	templates := GenerateTemplates()
	PrepareRoutesWithTemplates(router, templates)
	http.ListenAndServe(":8080", router)
}
