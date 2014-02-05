package models

import (
	"github.com/robmerrell/vtcboard/config"
	"labix.org/v2/mgo"
)

// MgoConnection holds a database connection and should be
// cloned frequently.
type MgoConnection struct {
	Session *mgo.Session
	DB      *mgo.Database
	DBName  string
}

// Close closes the database connection
func (m *MgoConnection) Close() {
	m.Session.Close()
}

// mainConnection is the default package level connection to the database.
// It should never be used directly, but cloned (which uses the same connection
// pool) and all clone need to call Close() when finished.
var mainConnection = &MgoConnection{}

// ConnectToDB connects the main connection to the database at a given server
// and also creates a reference to the given database.
func ConnectToDB(server, database string) error {
	// connect to the db server
	var err error
	mainConnection.Session, err = mgo.Dial(server)
	if err != nil {
		return err
	}

	// use a database
	mainConnection.DBName = database
	mainConnection.DB = mainConnection.Session.DB(database)
	return nil
}

// CloseDB closes the main database connection.
func CloseDB() {
	mainConnection.Close()
}

// CloneConnection clones the main database connection and returns it. Remember
// that cloned connections also need to be closed.
func CloneConnection() *MgoConnection {
	cloned := mainConnection.Session.Clone()
	return &MgoConnection{
		Session: cloned,
		DB:      cloned.DB(mainConnection.DBName),
		DBName:  mainConnection.DBName,
	}
}

// DropCollection drops all collections in the database. For this function to
// work the config environment must be set to "test".
func DropCollections() {
	if config.String("env") != "test" {
		panic("DropCollections only works in the test environment")
	}

	conn := CloneConnection()
	defer conn.Close()

	collections := []string{priceCollection, networkCollection, postCollection, averageCollection}
	for _, collection := range collections {
		conn.DB.C(collection).DropCollection()
	}
}
