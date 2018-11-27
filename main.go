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

var flagUser string
var flagPass string

type Document struct {
	ID   string
	Name string
	Size int
}

type DocumentDAO struct {
	ID   string
	Name string
	Size int
	Path string
}

var docs map[string]DocumentDAO

func main() {
	router := mux.NewRouter()
	docs = make(map[string]DocumentDAO)
	flagUser = "user"
	flagPass = "pass"
	router.HandleFunc("/documents", use(getDocuments, basicAuth)).Methods("GET")
	router.HandleFunc("/documents/{id}", use(getDocumentById, basicAuth)).Methods("GET")
	router.HandleFunc("/documents/download/{id}", use(serveDocuments, basicAuth)).Methods("GET")
	router.HandleFunc("/documents/{id}", use(deleteDocuments, basicAuth)).Methods("DELETE")
	router.HandleFunc("/documents", use(uploadDocument, basicAuth)).Methods("POST")
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
	f, err := os.OpenFile("./temp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	io.Copy(f, file)

}

func getDocumentById(w http.ResponseWriter, r *http.Request) {
	//var docs []Document
	docs = loadDocuments(docs)
	vars := mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")
	var doc DocumentDAO
	for _, v := range docs {
		if v.ID == vars["id"] {
			doc = v

		}
	}

	if documentInArray(vars["id"], docs) {
		json.NewEncoder(w).Encode(parseDocument(doc))
	} else {
		http.Error(w, "", http.StatusNotFound)

	}

}

func documentInArray(a string, list map[string]DocumentDAO) bool {
	for _, b := range list {
		if b.ID == a {
			return true
		}
	}
	return false
}

func getDocuments(w http.ResponseWriter, r *http.Request) {
	var docs map[string]DocumentDAO
	docs = make(map[string]DocumentDAO)
	docs = loadDocuments(docs)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(parseDocuments(docs))
	//json.NewEncoder(w).Encode(docs)

}

func loadDocuments(docs map[string]DocumentDAO) map[string]DocumentDAO {
	root := "./temp/"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.Name() != "temp" {
			id, error := hash_file_md5(path)
			if error == nil {
				//*docs = append(*docs,
				//	Document{ID: id, Name: info.Name(), Size: int(info.Size())})
				docs[id] = DocumentDAO{ID: id, Name: info.Name(), Size: int(info.Size()), Path: path}
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return docs
}

func deleteDocuments(w http.ResponseWriter, r *http.Request) {
	//var docs []Document
	docs = loadDocuments(docs)
	vars := mux.Vars(r)

	w.Header().Set("Content-Type", "application/json")

	if documentInArray(vars["id"], docs) {
		deleteDocument(vars["id"])
	} else {
		http.Error(w, "", http.StatusNotFound)

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

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, _ := r.BasicAuth()

		if flagUser != user || flagPass != pass {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized.", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}

func serveDocuments(w http.ResponseWriter, r *http.Request) {
	//var docs []Document
	docs = loadDocuments(docs)
	vars := mux.Vars(r)
	var docPath string
	//w.Header().Set("Content-Type", "application/octet-stream")
	//w.Header().Set("Content-Disposition", "attachment")

	if documentInArray(vars["id"], docs) {
		docPath = serveDocument(vars["id"])
		if docPath != "" {

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", "attachment; filename="+docs[vars["id"]].Name)
			http.ServeFile(w, r, docPath)
		} else {
			http.Error(w, "", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "", http.StatusNotFound)

	}

}

func serveDocument(docId string) string {
	root := "./temp/"
	var docPath string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.Name() != "temp" {
			id, error := hash_file_md5(path)
			if error == nil {
				if id == docId {
					docPath = path
				}
			}
		}
		return nil
	})

	return docPath

}

func parseDocuments(dao map[string]DocumentDAO) map[string]Document {
	var d map[string]Document
	d = make(map[string]Document)
	for _, data := range dao {
		d[data.ID] = Document{ID: data.ID, Name: data.Name, Size: data.Size}
	}
	return d
}

func parseDocument(data DocumentDAO) Document {
	return Document{ID: data.ID, Name: data.Name, Size: data.Size}

}
