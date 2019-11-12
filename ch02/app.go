package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"gopkg.in/validator.v2"
)

type App struct {
	Router      *mux.Router
	Middlewares *Middleware
	Config      *Env
}

type shortenReq struct {
	URL                 string `json:"url" validate: "nonzero"`
	ExpirationInMinutes int64  `json:"expiration_in_minutes" validate: "nonzero"`
}

type shortlinkResp struct {
	Shortlink string `json:"shortlink"`
}

func (a *App) Initialize(e *Env) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	a.Config = e
	a.Router = mux.NewRouter()
	a.Middlewares = &Middleware{}

	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	// a.Router.HandleFunc("/api/shorten", a.createShortlink).Methods("POST")
	// a.Router.HandleFunc("/api/info", a.getShortlinkInfo).Methods("GET")
	// a.Router.HandleFunc("/{shortlink:[a-zA-Z0-9]{1,11}}", a.redirect).Methods("GET")

	m := alice.New(a.Middlewares.LoggingHandler, a.Middlewares.RecoverHandler)

	a.Router.Handle("/api/shorten",
		m.ThenFunc(a.createShortlink)).Methods("POST")

	a.Router.Handle("/api/info",
		m.ThenFunc(a.getShortlinkInfo)).Methods("GET")

	a.Router.Handle("/{shortlink:[a-zA-Z0-9]{1,11}}",
		m.ThenFunc(a.redirect)).Methods("GET")
}

func (a *App) createShortlink(w http.ResponseWriter, r *http.Request) {
	var (
		req shortenReq
		err error
		s   string
	)
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("parse parameters failed %v", r.Body)})
		return
	}

	if err = validator.Validate(req); err != nil {
		respondWithError(w, StatusError{http.StatusBadRequest,
			fmt.Errorf("validate parameters failed %v", req)})
		return
	}

	defer r.Body.Close()

	// fmt.Printf("%v\n", req)
	if s, err = a.Config.S.Shorten(req.URL, req.ExpirationInMinutes); err != nil {
		respondWithError(w, err)
	} else {
		respondWithJSON(w, http.StatusCreated, shortlinkResp{Shortlink: s})
	}
}

func (a *App) getShortlinkInfo(w http.ResponseWriter, r *http.Request) {
	var (
		vals url.Values
		err  error
		s    string
		d    interface{}
	)

	vals = r.URL.Query()
	s = vals.Get("shortlink")

	// fmt.Printf("%v\n", s)
	if d, err = a.Config.S.ShortlinkInfo(s); err != nil {
		respondWithError(w, err)
	} else {
		respondWithJSON(w, http.StatusOK, d)
	}
}

func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// fmt.Printf("%s\n", vars["shortlink"])
	if u, err := a.Config.S.Unshorten(vars["shortlink"]); err != nil {
		respondWithError(w, err)
	} else {
		respondWithJSON(w, http.StatusTemporaryRedirect, u)
	}
}

// Run ..
func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

func respondWithError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case Error:
		log.Printf("HTTP %d - %s", e.Status(), e)
		respondWithJSON(w, e.Status(), e.Error())
	default:
		respondWithJSON(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	resp, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(resp)
}
