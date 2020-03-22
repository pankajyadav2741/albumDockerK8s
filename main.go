package main

import (
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	//"io/ioutil"
	"log"
	"net/http"
)

type Albums struct {
	Name string `json:"name"`
	Image []Image `json:"image"`
}

type Image struct {
	Name string `json:"name"`
}

var albums []Albums
var Session *gocql.Session

func init() {
	var err error
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "albumspace"
	Session, err = cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	fmt.Println("Cassandra init done")

	//TODO: Create Keyspace

	//USE cluster.Keyspace
	//TODO: Create TYPE
	//CREATE TYPE IF NOT EXISTS test.image3 ( imgname text);

	//TODO: Create Table
	//CREATE TABLE IF NOT EXISTS test.album3 (albname text PRIMARY KEY, images list<FROZEN <image3>>);

}

//Show all albums
func showAlbum(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Displaying album names:\n")
	//CQL Operation
	iter:=Session.Query("SELECT albname FROM albumtable;").Iter()
	var data string
	for iter.Scan(&data){
		json.NewEncoder(w).Encode(data)
	}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}
}

//Create a new album
func addAlbum(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","application/json")
	param := mux.Vars(r)
	if err:= Session.Query(`INSERT INTO albumtable (albname) VALUES (?) IF NOT EXISTS;`,param["album"]).Exec();err!=nil {
		fmt.Println(err)
	} else {
		fmt.Fprintf(w, "New album added")
	}
}

//Delete an existing album
func deleteAlbum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type","application/json")
	param := mux.Vars(r)
	//CQL Operation
	if err:= Session.Query(`DELETE FROM albumtable WHERE albname=? IF EXISTS;`,param["album"]).Exec();err!=nil {
		fmt.Println(err)
	} else {
		fmt.Fprintf(w, "Album deleted")
	}
}

//Show all images in an album
func showImagesInAlbum(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type","application/json")
	param := mux.Vars(r)
	iter:=Session.Query("SELECT imagelist FROM albumtable WHERE albname=?;",param["album"]).Iter()
	var data []string
	for iter.Scan(&data){
		json.NewEncoder(w).Encode(data)
	}
	if err := iter.Close(); err != nil {
		log.Fatal(err)
	}
}

//Show a particular image inside an album
func showImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type","application/json")
	param := mux.Vars(r)
	iter:=Session.Query("SELECT imagelist FROM albumtable WHERE albname='?';",param["image"]).Iter()
	var data []string
	for iter.Scan(&data){
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
	w.Header().Set("Content-Type","application/json")
	param := mux.Vars(r)
	//CQL Operation
	if err:= Session.Query(`UPDATE albumtable SET imagelist=imagelist+['?'] WHERE albname=?;`,param["image"],param["album"]).Exec();err!=nil {
		fmt.Println(err)
	} else {
		fmt.Fprintf(w, "New image added")
	}
}

//Delete an image in an album
func deleteImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type","application/json")
	param := mux.Vars(r)
	//CQL Operation
	if err:= Session.Query(`UPDATE albumtable SET imagelist=imagelist-['?'] WHERE albname=?;`,param["image"],param["album"]).Exec();err!=nil {
		fmt.Println(err)
	} else {
		fmt.Fprintf(w, "New image added")
	}
}

func main() {
	//Initialize Router
	myRouter := mux.NewRouter().StrictSlash(true)

	//Show all albums
	myRouter.HandleFunc("/",showAlbum).Methods(http.MethodGet)
	//Create a new album
	myRouter.HandleFunc("/{album}",addAlbum).Methods(http.MethodPost)
	//Delete an existing album
	myRouter.HandleFunc("/{album}",deleteAlbum).Methods(http.MethodDelete)

	//Show all images in an album
	myRouter.HandleFunc("/{album}",showImagesInAlbum).Methods(http.MethodGet)
	//Show a particular image inside an album
	myRouter.HandleFunc("/{album}/{image}",showImage).Methods(http.MethodGet)
	//Create an image in an album
	myRouter.HandleFunc("/{album}/{image}",addImage).Methods(http.MethodPost)
	//Delete an image in an album
	myRouter.HandleFunc("/{album}/{image}",deleteImage).Methods(http.MethodDelete)
	log.Fatal(http.ListenAndServe(":8085",myRouter))
}
