package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"

	"github.com/jmoiron/sqlx"
	_ "gopkg.in/cq.v1"
)

// Database wraps a sqlx.DB in order to expose a custom, simple API.
type Database struct {
	*sqlx.DB
}

// NewDatabase connects to the Neo4j database and returns
// a pointer to the newly create Database.
func NewDatabase() *Database {
	db, err := sqlx.Connect("neo4j-cypher", config.Neo4jAddr)
	if err != nil {
		panic(err)
	}
	return &Database{db}
}

// GetUser fetches a user by id.
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

// GetItem fetches an item by id.
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

// UpsertUser creates or updates a User.
func (db *Database) UpsertUser(u *User) error {
	if u.Id == "" {
		u.Id = CreateId()
	}
	cypher := `MERGE (u:User {id: {0}}) SET u.name = {1}`

	_, err := db.Exec(cypher, u.Id, u.Name)
	return err
}

// UpsertItem creates or updates an Item.
func (db *Database) UpsertItem(i *Item) error {
	if i.Id == "" {
		i.Id = CreateId()
	}
	cypher := `MERGE (i:Item {id: {0}}) SET i.name = {1}, i.categories = {2}`

	_, err := db.Exec(cypher, i.Id, i.Name, i.Categories)
	return err
}

// UpsertLink creates or updates a Link.
func (db *Database) UpsertLink(l *Link) error {
	if l.Id == "" {
		l.Id = CreateId()
	}
	cypher := `MATCH (u:User {id:{0}}), (i:Item {id:{1}})
		MERGE (u)-[l:LINKED {id:{2}}]->(i)
		SET l.type = {3}, l.strength = {4}`

	_, err := db.Exec(cypher, l.UserId, l.ItemId, l.Id, l.Type, l.Score)
	return err
}

// Recommend retrieves recommended items for the given user, optionally
// filtering by category. Results are ordered by recommendation strength.
func (db *Database) Recommend(uid, linkType string, category []string) ([]Recommendation, error) {
	// Store binding vars in a slice
	args := []interface{}{}
	argPos := func() string {
		return fmt.Sprintf("{%d}", len(args)-1)
	}

	args = append(args, uid)
	where := " WHERE u.id = " + argPos()
	whereCat := ""                                 // Defaults to any category
	whereType := " AND NOT (u)-[:LINKED]->(item2)" // Defaults to any type
	if len(category) > 0 {
		args = append(args, category)
		whereCat = " AND ANY (x IN " + argPos() + " WHERE x in item2.categories)" // Any category match will be enough
	}
	if linkType != "" {
		args = append(args, linkType)
		whereType = " AND l1.type = " + argPos() + " AND l2.type = " + argPos() + " AND l3.type = " + argPos() +
			" AND NOT (u)-[:LINKED {type:" + argPos() + "}]->(item2)"
	}

	cypher := `MATCH (u:User)-[l1:LINKED]->(item1:Item)<-[l2:LINKED]-(u2:User),
		(u2)-[l3:LINKED]->(item2:Item)` +
		where + whereCat + whereType +
		` RETURN item2.id, item2.name, count(distinct l3) as frequency
		ORDER BY frequency DESC`

	rows, err := db.Query(cypher, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	resp := []Recommendation{}
	for rows.Next() {
		var rec Recommendation
		err = rows.Scan(&rec.Item.Id, &rec.Item.Name, &rec.Strength)
		if err != nil {
			return nil, err
		}
		resp = append(resp, rec)
	}
	return resp, nil
}

// CreateId returns an hexadecimal, 8-byte, random ID.
func CreateId() string {
	id := make([]byte, 8)
	io.ReadFull(rand.Reader, id)
	return fmt.Sprintf("%x", id)
}
