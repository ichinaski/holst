package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"

	"github.com/jmoiron/sqlx"
	_ "gopkg.in/cq.v1"
)

type Database struct {
	*sqlx.DB
}

func NewDatabase() *Database {
	db, err := sqlx.Connect("neo4j-cypher", neo4jURL)
	if err != nil {
		panic(err)
	}
	return &Database{db}
}

func (db *Database) GetUser(id string) *User {
	cypher := `MATCH (u:User)
			 	WHERE u.id = {0}
			 	RETURN u.id as id, u.name as name
			 	LIMIT 1`

	user := &User{}
	err := db.Get(user, cypher, id)

	if err != nil {
		if err != sql.ErrNoRows {
			Logger.Println(err)
		}
		return nil
	}
	return user
}

func (db *Database) GetItem(id string) *Item {
	cypher := `MATCH (i:Item)
			 	WHERE i.id = {0}
			 	RETURN i.id as id, i.name as name
			 	LIMIT 1`

	item := &Item{}
	err := db.Get(item, cypher, id)

	if err != nil {
		if err != sql.ErrNoRows {
			Logger.Println(err)
		}
		return nil
	}
	return item
}

func (db *Database) UpsertUser(u *User) error {
	if u.Id == "" {
		u.Id = CreateId()
	}
	cypher := `MERGE (u:User {id: {0}})
				SET u.name = {1}`

	_, err := db.Exec(cypher, u.Id, u.Name)
	return err
}

func (db *Database) UpsertItem(i *Item) error {
	if i.Id == "" {
		i.Id = CreateId()
	}
	cypher := `MERGE (i:Item {id: {0}})
				SET i.name = {1}, i.categories = {2}`

	_, err := db.Exec(cypher, i.Id, i.Name, i.Categories)
	return err
}

func (db *Database) UpsertLink(l *Link) error {
	if l.Id == "" {
		l.Id = CreateId()
	}
	cypher := `MATCH (u:User {id:{0}}), (i:Item {id:{1}})
				MERGE (u)-[l:LINKED {id:{2}}]->(i)
				SET l.type = {3}`

	_, err := db.Exec(cypher, l.UserId, l.ItemId, l.Id, l.Type)
	return err
}

func CreateId() string {
	// TODO: Use UUIDs instead
	id := make([]byte, 8)
	io.ReadFull(rand.Reader, id)
	return fmt.Sprintf("%x", id)
}
