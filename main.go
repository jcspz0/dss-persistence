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
	router.HandleFunc("/documents/{id}", getDocumentById).Methods("GET")
	router.HandleFunc("/documents/{id}", deleteDocuments).Methods("DELETE")
	router.HandleFunc("/documents", uploadDocument).Methods("POST")
	log.Fatal(http.ListenAndServe(":9001", router))
}

func uploadDocument(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	//fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.OpenFile("./temp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(f, file)

}

func getDocumentById(w http.ResponseWriter, r *http.Request) {
	var docs []Document
	loadDocuments(&docs)
	vars := mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")
	var doc Document
	for _, v := range docs {
		if v.ID == vars["id"] {
			doc = v

		}
	}

	if documentInArray(vars["id"], docs) {
		json.NewEncoder(w).Encode(doc)
	} else {
		http.Error(w, "", http.StatusNotFound)

	}

}

func documentInArray(a string, list []Document) bool {
	for _, b := range list {
		if b.ID == a {
			return true
		}
	}
	return false
}

func getDocuments(w http.ResponseWriter, r *http.Request) {
	var docs []Document
	loadDocuments(&docs)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(docs)

}

func deleteDocuments(w http.ResponseWriter, r *http.Request) {
	var docs []Document
	loadDocuments(&docs)
	vars := mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	if documentInArray(vars["id"], docs) {
		deleteDocument(vars["id"])
	} else {
		http.Error(w, "", http.StatusNotFound)

	}

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

func deleteDocument(docId string) bool {
	root := "./temp/"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.Name() != "temp" {
			id, error := hash_file_md5(path)
			if error == nil {
				if id == docId {
					os.Remove(path)
				}
			}
		}
		return nil
	})
	if err != nil {
		return true
	} else {
		return false
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
