package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
)

type Albums struct {
	Name  string  `json:"name"`
	Image []Image `json:"image"`
}

type Image struct {
	Name string `json:"name"`
}

var albums []Albums
var Session *gocql.Session

func init() {
	var host string
	err := envconfig.Process("DB_HOST", host)
	if err != nil {
		log.Fatal(err.Error())
	}

	cluster := gocql.NewCluster(host)
	cluster.Keyspace = "albumspace"
	Session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	fmt.Println("Cassandra init done")

	//Create Keyspace
	if err := Session.Query(`CREATE KEYSPACE ? IF NOT EXISTS WITH replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };`, cluster.Keyspace).Exec(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Fprintf(w, "")
	}

	//Use Keyspace
	if err := Session.Query(`USE ?;`, cluster.Keyspace).Exec(); err != nil {
		fmt.Println(err)
	}

	//Create Table albumtable
	if err := Session.Query(`CREATE TABLE IF NOT EXISTS albumtable(albname TEXT PRIMARY KEY, imagelist LIST<TEXT>);`).Exec(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Fprintf(w, "Table albumtable created")
	}
}

//Show all albums
func showAlbum(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Displaying album names:\n")
	//CQL Operation
	iter := Session.Query("SELECT albname FROM albumtable;").Iter()
	var data string
	for iter.Scan(&data) {
		json.NewEncoder(w).Encode(data)
	}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}
}

//Create a new album
func addAlbum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	param := mux.Vars(r)
	if err := Session.Query(`INSERT INTO albumtable (albname) VALUES (?) IF NOT EXISTS;`, param["album"]).Exec(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Fprintf(w, "New album added")
	}
}

//Delete an existing album
func deleteAlbum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	param := mux.Vars(r)
	//CQL Operation
	if err := Session.Query(`DELETE FROM albumtable WHERE albname=? IF EXISTS;`, param["album"]).Exec(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Fprintf(w, "Album deleted")
	}
}

//Show all images in an album
func showImagesInAlbum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	param := mux.Vars(r)
	iter := Session.Query("SELECT imagelist FROM albumtable WHERE albname=?;", param["album"]).Iter()
	var data []string
	for iter.Scan(&data) {
		json.NewEncoder(w).Encode(data)
	}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}
}

//Show a particular image inside an album
func showImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	param := mux.Vars(r)
	iter := Session.Query("SELECT imagelist FROM albumtable WHERE albname='?';", param["image"]).Iter()
	var data []string
	for iter.Scan(&data) {
		for _, img := range data {
			if img == param["image"] {
				json.NewEncoder(w).Encode(img)
			}
		}
	}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}
}

//Create an image in an album
func addImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	param := mux.Vars(r)
	//CQL Operation
	if err := Session.Query(`UPDATE albumtable SET imagelist=imagelist+['?'] WHERE albname=?;`, param["image"], param["album"]).Exec(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Fprintf(w, "New image added")
	}
}

//Delete an image in an album
func deleteImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	param := mux.Vars(r)
	//CQL Operation
	if err := Session.Query(`UPDATE albumtable SET imagelist=imagelist-['?'] WHERE albname=?;`, param["image"], param["album"]).Exec(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Fprintf(w, "Image deleted")
	}
}

func main() {
	//Initialize Router
	myRouter := mux.NewRouter().StrictSlash(true)

	//Show all albums
	myRouter.HandleFunc("/", showAlbum).Methods(http.MethodGet)
	//Create a new album
	myRouter.HandleFunc("/{album}", addAlbum).Methods(http.MethodPost)
	//Delete an existing album
	myRouter.HandleFunc("/{album}", deleteAlbum).Methods(http.MethodDelete)

	//Show all images in an album
	myRouter.HandleFunc("/{album}", showImagesInAlbum).Methods(http.MethodGet)
	//Show a particular image inside an album
	myRouter.HandleFunc("/{album}/{image}", showImage).Methods(http.MethodGet)
	//Create an image in an album
	myRouter.HandleFunc("/{album}/{image}", addImage).Methods(http.MethodPost)
	//Delete an image in an album
	myRouter.HandleFunc("/{album}/{image}", deleteImage).Methods(http.MethodDelete)
	log.Fatal(http.ListenAndServe(":8085", myRouter))
}
