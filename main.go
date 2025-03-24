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
	Sender string
	Text   string
}

type Messages struct {
	Messages []Message
}

func NewMessages() Messages {
	return Messages{
		make([]Message, 0, 10),
	}
}

func (m *Messages) AddMessage(sender, text string) {
	m.Messages = append(m.Messages, Message{sender, text})
}

func (m *Messages) GetMessages() []Message {
	return m.Messages
}

func (m *Messages) AddUserMessage(msg string) {
	m.AddMessage("User", msg)
}

func (m *Messages) AddBotMessage(msg string) {
	m.AddMessage("Bot", msg)
}

func GenerateTemplates() *template.Template {
	t := template.Must(template.ParseGlob("templates/layout/*.gohtml"))
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

	msgs := NewMessages()

	return func(w http.ResponseWriter, r *http.Request) {
		if e := r.ParseForm(); e != nil {
			fmt.Fprintf(os.Stderr, "could not parse form")
			writeHeaders(w, http.StatusBadRequest)
			return
		}

		msg := r.FormValue("message")

		msgs.AddUserMessage(msg)

		geminiResp, err := gemini.GenerateResponse(msg)

		if err == nil {
			msgs.AddBotMessage(geminiResp.Text)
		} else {
			msgs.AddBotMessage("Cannot talk to Gemini at the moment. 🥲")
		}

		writeHeaders(w, http.StatusOK)
		err = templates.ExecuteTemplate(w, "index.gohtml", msgs)

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
