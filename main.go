package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

var db, err = badger.Open(badger.DefaultOptions("./"))

type (
	Message struct {
		Code int
		Msg  string
	}

	JSON struct {
		Key   string
		Value string
	}
)

func index(w http.ResponseWriter, r *http.Request) {
	send := Message{200, "Hi there"}
	js, _ := json.Marshal(send)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(js)
}

// store our key value, a bit like json_encode and decode
func storeKV(w http.ResponseWriter, r *http.Request) {

	var kvPair JSON

	err0 := json.NewDecoder(r.Body).Decode(&kvPair)
	if err0 != nil {
		handle(w, err0)
		return
	}

	txn := db.NewTransaction(true)
	defer txn.Discard()

	err := txn.Set([]byte(kvPair.Key), []byte(kvPair.Value))
	if err != nil {
		log.Fatal(err)
	}

	// Commit the transaction and check for error.
	if err := txn.Commit(); err != nil {
		log.Fatal(err)
	}

	send := JSON{kvPair.Key, kvPair.Value}

	js, err3 := json.Marshal(send)
	if err3 != nil {
		handle(w, err3)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(js)
}

func getValue(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(vars["key"]))
		if err != nil {
			send(w, "404", "Not Found")
		}

		var valCopy []byte

		valCopy, err = item.ValueCopy(nil)
		if err != nil {
			log.Fatal(err)
		}
		send(w, vars["key"], string(valCopy))
		return nil
	})

}

func handle(w http.ResponseWriter, err error) {
	send := Message{500, err.Error()}
	js, _ := json.Marshal(send)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(500)
	w.Write(js)
}

func send(w http.ResponseWriter, key string, value string) {
	format := JSON{key, value}
	js, _ := json.Marshal(format)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(js)
}

func main() {

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := ":" + os.Getenv("PORT")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", index)
	router.HandleFunc("/get/{key}", getValue).Methods("GET")
	router.HandleFunc("/store", storeKV).Methods("POST")

	log.Fatal(http.ListenAndServe(port, router))
}
