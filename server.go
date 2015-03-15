package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "gopkg.in/cq.v1"
)

const neo4jURL = "http://localhost:7474"

var Logger = log.New(os.Stdout, "  ", log.LstdFlags|log.Lshortfile)
var (
	ErrNotFound     = errors.New("Not Found")
	ErrBadRequest   = errors.New("Bad Request")
	ErrUnauthorized = errors.New("Unauthorized")
)

type Context struct {
	db *Database
}

func NewContext(r *http.Request) *Context {
	db := NewDatabase()
	return &Context{db}
}

func handler(f func(ctx *Context, w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Logger.Printf("%v: %v\n", r.Method, r.URL.Path)

		ctx := NewContext(r)
		defer func() {
			if err := ctx.db.Close(); err != nil {
				Logger.Println(err)
			}
		}()

		err := f(ctx, w, r)

		if err == nil {
			return
		}

		Logger.Println("Error: ", err)
		switch err {
		case ErrNotFound:
			http.Error(w, "Not Found", http.StatusNotFound)
		case ErrBadRequest:
			http.Error(w, "Bad Request", http.StatusBadRequest)
		case ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		default:
			http.Error(w, "Oops!", http.StatusInternalServerError)
		}

	}
}

func userHandler(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	switch {
	case r.Method == "POST":
		user := &User{}
		if err := json.NewDecoder(r.Body).Decode(user); err != nil {
			Logger.Println(err)
			return ErrBadRequest
		}

		err := ctx.db.UpsertUser(user)
		if err != nil {
			return err
		}

		return writeJSON(w, user, http.StatusCreated)
	case r.Method == "GET":
		id := mux.Vars(r)["id"]
		if id == "" {
			return ErrBadRequest
		}

		user := ctx.db.GetUser(id)
		if user == nil {
			return ErrNotFound
		}
		return writeJSON(w, user, http.StatusOK)
	}
	return ErrBadRequest
}

func itemHandler(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	switch {
	case r.Method == "POST":
		item := &Item{}
		if err := json.NewDecoder(r.Body).Decode(item); err != nil {
			Logger.Println(err)
			return ErrBadRequest
		}

		if err := ctx.db.UpsertItem(item); err != nil {
			return err
		}
		return writeJSON(w, item, http.StatusCreated)
	case r.Method == "GET":
		id := mux.Vars(r)["id"]
		if id == "" {
			return ErrBadRequest
		}

		item := ctx.db.GetItem(id)
		if item == nil {
			return ErrNotFound
		}

		return writeJSON(w, item, http.StatusOK)
	}
	return ErrBadRequest
}

func linkHandler(ctx *Context, w http.ResponseWriter, r *http.Request) error {
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
		err := ctx.db.UpsertLink(link)
		if err != nil {
			return err
		}

		return writeJSON(w, link, http.StatusCreated)
	}
	return ErrBadRequest
}

// recommendHandler will manage item recommendations. It currently reads the user id and
// categories for the recommended items. Matching items must fulfil *any* category
func recommendHandler(ctx *Context, w http.ResponseWriter, r *http.Request) error {
	uid := mux.Vars(r)["uid"]
	if uid == "" {
		return ErrBadRequest
	}

	r.ParseForm()
	categories := r.Form["category"]

	// Store binding vars in a slice
	args := []interface{}{}
	argPos := func() string {
		return strconv.Itoa(len(args) - 1) // Current var position (string)
	}

	args = append(args, uid)
	where := "WHERE u.id = {" + argPos() + "}"
	if len(categories) > 0 {
		//where = where + " AND ALL (x IN {1} WHERE x in item2.categories)"
		args = append(args, categories)
		where = where + " AND ANY (x IN {" + argPos() + "} WHERE x in item2.categories)"
	}

	cypher := `MATCH (u:User)-[:LINKED]->(item1:Item)<-[:LINKED]-(u2:User),
		(u2)-[l:LINKED]->(item2:Item)` +
		where +
		`AND NOT (u)-[:LINKED]->(item2)
		RETURN item2.id, item2.name, count(distinct l) as frequency
		ORDER BY frequency DESC`

	rows, err := ctx.db.Query(cypher, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	resp := []Recommendation{}
	for rows.Next() {
		var rec Recommendation
		err = rows.Scan(&rec.Item.Id, &rec.Item.Name, &rec.Strength)
		if err != nil {
			return err
		}
		resp = append(resp, rec)
	}

	return writeJSON(w, resp, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, value interface{}, status int) error {
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=UTF8")
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.Write(body)
	return nil
}

func main() {
	// TODO: Create uniqueness constraints in Graph (user id, item id)

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
