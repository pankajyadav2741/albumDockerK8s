package main

import (
	"encoding/json"
	"fmt"
	"os"
	"log"
	"net/http"
	"os/signal"
	"time"
	"context"
	"syscall"
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

type Conf struct {
	DbHost string `envconfig:"DB_HOST"`
}

var albums []Albums
var cluster *gocql.ClusterConfig

func init() {
	db := &Conf{}
	err := envconfig.Process("", db)
	if err != nil {
		fmt.Println("Error in envconfig")
		log.Fatal(err.Error())
	}
	fmt.Println("DB HOST", db.DbHost)
	cluster = gocql.NewCluster(db.DbHost)
	cluster.Keyspace = "system"
	cluster.Timeout  = time.Second * 20
	cluster.ConnectTimeout  = time.Second * 20
	//cluster.DisableInitialHostLookup = true

	Session, err := cluster.CreateSession()
	defer Session.Close()
	if err != nil {
		fmt.Println("Create session failed")
		panic(err)
	}
	fmt.Println("Cassandra init done")

	//Create Keyspace
	if err := Session.Query(`CREATE KEYSPACE IF NOT EXISTS albumspace WITH replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };`).Exec(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Keyspace created")
	}
	
	cluster.Keyspace = "albumspace"
	session, err := cluster.CreateSession()
	defer session.Close()
	if err != nil {
		log.Fatal("createSession:", err)
	}
	
	//Create Table albumtable
	if err := session.Query(`CREATE TABLE IF NOT EXISTS albumtable(albname TEXT PRIMARY KEY, imagelist LIST<TEXT>);`).Exec(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Table albumtable created")
	}
}

//Show all albums
func showAlbum(w http.ResponseWriter, r *http.Request) {
	Session, err := cluster.CreateSession()
	defer Session.Close()
	if err != nil {
		fmt.Println("Create session failed")
		panic(err)
	}
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
	Session, err := cluster.CreateSession()
	defer Session.Close()
	if err != nil {
		fmt.Println("Create session failed")
		panic(err)
	}
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
	Session, err := cluster.CreateSession()
	defer Session.Close()
	if err != nil {
		fmt.Println("Create session failed")
		panic(err)
	}
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
	Session, err := cluster.CreateSession()
	defer Session.Close()
	if err != nil {
		fmt.Println("Create session failed")
		panic(err)
	}
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
	Session, err := cluster.CreateSession()
	defer Session.Close()
	if err != nil {
		fmt.Println("Create session failed")
		panic(err)
	}
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
	Session, err := cluster.CreateSession()
	defer Session.Close()
	if err != nil {
		fmt.Println("Create session failed")
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	param := mux.Vars(r)
	//CQL Operation
	if err := Session.Query(`UPDATE albumtable SET imagelist=imagelist+ ? WHERE albname=?;`, []string{param["image"]}, param["album"]).Exec(); err != nil {
		fmt.Println(err)
	} else {
		fmt.Fprintf(w, "New image added")
	}
}

//Delete an image in an album
func deleteImage(w http.ResponseWriter, r *http.Request) {
	Session, err := cluster.CreateSession()
	defer Session.Close()
	if err != nil {
		fmt.Println("Create session failed")
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	param := mux.Vars(r)
	//CQL Operation
	if err := Session.Query(`UPDATE albumtable SET imagelist=imagelist-? WHERE albname=?;`, []string{param["image"]}, param["album"]).Exec(); err != nil {
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
	
	srv := &http.Server{
		Handler:      myRouter,
		Addr:         ":5000",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start Server
	go func() {
		log.Println("Starting Server")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful Shutdown
	waitForShutdown(srv)
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}
