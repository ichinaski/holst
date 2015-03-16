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

const neo4jURL = "http://localhost:7474"

var (
	db     *Database
	Logger = log.New(os.Stdout, "  ", log.LstdFlags|log.Lshortfile)
)
var (
	ErrNotFound     = errors.New("Not Found")
	ErrBadRequest   = errors.New("Bad Request")
	ErrUnauthorized = errors.New("Unauthorized")
)

func handler(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rv := recover(); rv != nil {
				Logger.Println("Error: handler panic!")
				writeAPIError(w, http.StatusInternalServerError)
			}
		}()
		Logger.Printf("%v: %v\n", r.Method, r.URL.Path)

		// TODO: Implement optional authentication param. Transparently handle the auth before calling the handler.
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

		/*
			// This query needs to match *all* nodes, therefore is not valid
			cypher := `MERGE (u:User {id:{0}, name:{1}})
						MERGE (i:Item {id:{2}, name:{3}})
						MERGE (u)-[r:LINKED {type:{4}}]->(i)` // FIXME: No return values?
			rows, err := ctx.db.Query(cypher, link.User.Id, link.User.Name, link.Item.Id, link.Item.Name, link.Type)
		*/
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
	vars := mux.Vars(r)
	uid, linkType := vars["uid"], vars["type"]
	if uid == "" {
		return ErrBadRequest
	}

	r.ParseForm()
	category := r.Form["category"]

	recs, err := db.Recommend(uid, linkType, category)
	if err != nil {
		return err
	}
	return writeJSON(w, recs, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, value interface{}, status int) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=UTF8")
	return json.NewEncoder(w).Encode(value)
}

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

func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

func String(s string) *string {
	p := new(string)
	*p = s
	return p
}
