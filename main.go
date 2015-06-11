package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/golang/gddo/httputil"
	"github.com/gorilla/mux"
	_ "gopkg.in/cq.v1"
)

var (
	db     *Database
	config = loadConfig()
	Logger = log.New(os.Stdout, "  ", log.LstdFlags|log.Lshortfile)
)

var (
	ErrNotFound     = errors.New("Not Found")
	ErrBadRequest   = errors.New("Bad Request")
	ErrUnauthorized = errors.New("Unauthorized")
)

func main() {
	// TODO: Create uniqueness constraints in Graph (user id, item id)
	db = NewDatabase()

	r := mux.NewRouter()
	r.HandleFunc("/user", handler(userHandler)).Methods("POST")
	r.HandleFunc("/user/{id}", handler(userHandler)).Methods("GET")
	r.HandleFunc("/item", handler(itemHandler)).Methods("POST")
	r.HandleFunc("/item/{id}", handler(itemHandler)).Methods("GET")
	r.HandleFunc("/link", handler(linkHandler)).Methods("POST")

	r.HandleFunc("/recommend/{uid}", handler(recommendHandler)).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

// handler wraps a custom handler and returns a standard http.HandleFunc,
// managing common error situations and JSON responses.
func handler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rv := recover(); rv != nil {
				Logger.Println("Error: handler panic!")
				writeAPIError(w, http.StatusInternalServerError)
			}
		}()
		Logger.Printf("%v: %v\n", r.Method, r.URL.Path)

		// Handle authentication.
		user, pass, ok := r.BasicAuth()
		if !ok || user != config.HttpUsername || pass != config.HttpPassword {
			writeAPIError(w, http.StatusUnauthorized)
			return
		}

		var wb httputil.ResponseBuffer
		err := f(&wb, r)
		if err == nil {
			wb.WriteTo(w)
			return
		}

		Logger.Println("Error: ", err)
		switch err {
		case ErrNotFound:
			writeAPIError(w, http.StatusNotFound)
		case ErrBadRequest:
			writeAPIError(w, http.StatusBadRequest)
		case ErrUnauthorized:
			writeAPIError(w, http.StatusUnauthorized)
		default:
			writeAPIError(w, http.StatusInternalServerError)
		}

	}
}

// userHandler manages User creation and updates (POST), as well
// as User retrievals (GET).
func userHandler(w http.ResponseWriter, r *http.Request) error {
	switch {
	case r.Method == "POST":
		user := &User{}
		if err := json.NewDecoder(r.Body).Decode(user); err != nil {
			Logger.Println(err)
			return ErrBadRequest
		}

		err := db.UpsertUser(user)
		if err != nil {
			return err
		}

		return writeJSON(w, user, http.StatusCreated)
	case r.Method == "GET":
		id := mux.Vars(r)["id"]
		if id == "" {
			return ErrBadRequest
		}

		user := db.GetUser(id)
		if user == nil {
			return ErrNotFound
		}
		return writeJSON(w, user, http.StatusOK)
	}
	return ErrBadRequest
}

// itemHandler manages Item creation and updates (POST), as well
// as User retrievals (GET).
func itemHandler(w http.ResponseWriter, r *http.Request) error {
	switch {
	case r.Method == "POST":
		item := &Item{}
		if err := json.NewDecoder(r.Body).Decode(item); err != nil {
			Logger.Println(err)
			return ErrBadRequest
		}

		if err := db.UpsertItem(item); err != nil {
			return err
		}
		return writeJSON(w, item, http.StatusCreated)
	case r.Method == "GET":
		id := mux.Vars(r)["id"]
		if id == "" {
			return ErrBadRequest
		}

		item := db.GetItem(id)
		if item == nil {
			return ErrNotFound
		}

		return writeJSON(w, item, http.StatusOK)
	}
	return ErrBadRequest
}

// linkHandler manages Link creation and updates, returning an error
// if an error is encountered.
func linkHandler(w http.ResponseWriter, r *http.Request) error {
	switch {
	case r.Method == "POST":
		// For more complex JSON decodings: http://talks.golang.org/2015/json.slide#21
		link := &Link{}
		if err := json.NewDecoder(r.Body).Decode(link); err != nil {
			return err
		}
		if link.UserId == "" || link.ItemId == "" {
			return ErrBadRequest
		}

		err := db.UpsertLink(link)
		if err != nil {
			return err
		}

		return writeJSON(w, link, http.StatusCreated)
	}
	return ErrBadRequest
}

// recommendHandler will manage item recommendations. It currently reads the user id and
// categories for the recommended items. Matching items must fulfil *any* category
func recommendHandler(w http.ResponseWriter, r *http.Request) error {
	uid := mux.Vars(r)["uid"]
	if uid == "" {
		return ErrBadRequest
	}

	r.ParseForm()
	category := r.Form["category"]
	linkType := ""
	if len(r.Form["type"]) > 0 {
		linkType = r.Form["type"][0]
	}

	recs, err := db.Recommend(uid, linkType, category)
	if err != nil {
		return err
	}
	return writeJSON(w, recs, http.StatusOK)
}

// writeJSON writes the given value to the http response writer,
// with the appropriate status code and standard headers.
func writeJSON(w http.ResponseWriter, value interface{}, status int) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=UTF8")
	return json.NewEncoder(w).Encode(value)
}

// writeAPIError writes the given error to the http response writer,
// with the appropriate status code and standard headers.
func writeAPIError(w http.ResponseWriter, status int) {
	var data struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	data.Error.Message = http.StatusText(status)
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=UTF8")
	json.NewEncoder(w).Encode(&data)
}
