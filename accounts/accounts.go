package accounts

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
)

var (
    createTmpl = template.Must(template.ParseFiles("accounts/create.html"))
)

func create(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        createTmpl.Execute(w, nil)
        return
    }
    email := r.FormValue("email")
    name := r.FormValue("name")
    password := r.FormValue("password1")
    if password != r.FormValue("password2") {
        log.Fatal("TODO: mismatched pswd")
        return
    }
    fmt.Fprintf(w, "%s / %s / %s", email, name, password)
}

func RegisterHandlers(r *mux.Router) {
	r.HandleFunc("/create", http.HandlerFunc(create))
}
