package database

import (
	"fmt"

	"github.com/Skarlso/google-oauth-go-sample/structs"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// MongoDBConnection Encapsulates a connection to a database.
type MongoDBConnection struct {
}

// SaveUser register a user so we know that we saw that user already.
func (mdb *MongoDBConnection) SaveUser(u *structs.User) error {
	session, err := mdb.GetSession()
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	defer session.Close()

	if _, err := mdb.LoadUser(u.Email); err == nil {
		return fmt.Errorf("user already exists")
	}
	c := session.DB("webadventure").C("users")
	return c.Insert(u)
}

// LoadUser get data from a user.
func (mdb *MongoDBConnection) LoadUser(Email string) (*structs.User, error) {
	session, err := mdb.GetSession()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	defer session.Close()

	result := &structs.User{}
	c := session.DB("webadventure").C("users")
	err = c.Find(bson.M{"email": Email}).One(&result)

	return result, err
}

// GetSession return a new session if there is no previous one.
func (mdb *MongoDBConnection) GetSession() (*mgo.Session, error) {
	session, err := mgo.Dial("localhost")
	if err != nil {
		return nil, fmt.Errorf("failed to open session: %w", err)
	}
	session.SetMode(mgo.Monotonic, true)

	return session, nil
}
