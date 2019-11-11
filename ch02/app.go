package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
}

type shortenReq struct {
	URL                 string `json:"url" validate: "nonzero"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes" validate: "nonzero"`
}

type shortlinkResp struct {
	Shortlink string `json:"shortlink"`
}

func (a *App) Initialize() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/api/shorten", a.createShortlink).Methods("POST")
	a.Router.HandleFunc("/api/info", a.getShortlinkInfo).Methods("GET")
	a.Router.HandleFunc("/{shortlink:[a-zA-Z0-9]{1,11}}", a.redirect).Methods("GET")
}

func (a *App) createShortlink(w http.ResponseWriter, r *http.Request) {
	var (
		req shortenReq
		err error
	)
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		// fmt.Println("???")
		respondWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("parse parameters failed %v", r.Body)})
		return
	}

	// if err = validator.Validate(req); err != nil {
	// 	respondWithError(w, StatusError{http.StatusBadRequest,
	// 		fmt.Errorf("validate parameters failed %v", requrl)})
	// 	return
	// }

	defer r.Body.Close()

	fmt.Printf("%v\n", req)
}

func (a *App) getShortlinkInfo(w http.ResponseWriter, r *http.Request) {
	var (
		vals url.Values
		// err  error
		s string
	)

	vals = r.URL.Query()
	s = vals.Get("shortlink")

	fmt.Printf("%v\n", s)
}

func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	fmt.Printf("%s\n", vars["shortlink"])
}

// Run ..
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func respondWithError(w http.ResponseWriter, err error) {
	// fmt.Println("==")
	switch e := err.(type) {
	case Error:
		// fmt.Println("----")
		log.Printf("HTTP %d - %s", e.Status(), e)
		respondWithJSON(w, e.Status(), e.Error())
	default:
		fmt.Println("----2")
		respondWithJSON(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	resp, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(resp)
}
