package main

import (
	"fmt"
	"log"

	"gopkg.in/mgo.v2"
)

var getDb func() (*mgo.Database, *mgo.Session)

func initDb(dbName, dbHost, dbPort, credentials string) {
	if len(credentials) > 1 {
		dbHost = credentials + "@" + dbHost
	}
	url := fmt.Sprintf("mongodb://%s:%s/%s", dbHost, dbPort, dbName)
	mgoSession, err := mgo.Dial(url)
	checkErrorAndExit(err, 1)

	log.Println("Dialed:", url)

	getDb = func() (*mgo.Database, *mgo.Session) {
		s := mgoSession.Clone()
		return s.DB(dbName), s
	}
}
