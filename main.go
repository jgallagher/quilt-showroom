package main

import (
	"github.com/gorilla/mux"
	"github.com/jgallagher/dbproject/accounts"
	"log"
	"net/http"
)

func main() {
	r := mux.NewRouter()
	accounts.RegisterHandlers(r.PathPrefix("/account").Subrouter())
    //accounts.RegisterHandlers(r)
	http.Handle("/", r)
	err := http.ListenAndServe("127.0.0.1:6161", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
