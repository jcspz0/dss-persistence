package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

type Document struct {
	ID   string
	Name string
	Size int
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/documents", getDocuments).Methods("GET")
	log.Fatal(http.ListenAndServe(":9000", router))
}

func getDocuments(w http.ResponseWriter, r *http.Request) {
	var docs []Document
	loadDocuments(&docs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

func loadDocuments(docs *[]Document) {
	root := "./temp/"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.Name() != "temp" {
			id, error := hash_file_md5(path)
			if error == nil {
				*docs = append(*docs,
					Document{ID: id, Name: info.Name(), Size: int(info.Size())})
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func hash_file_md5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}
