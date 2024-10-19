package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
    "time"
)

var urlStorage = struct {
    sync.RWMutex
    mapping map[string]string
}{mapping: make(map[string]string)}

func generateShortUrl(lenght int) string {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
    b := make([]byte, lenght)
    for i := range b {
        b[i] = charset[seededRand.Intn(len(charset))]
    }
    return string(b)
}

func originUrlHandler(res http.ResponseWriter, req *http.Request) {

    if req.Method != http.MethodPost {
        http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
        return 
    }

    body, err := io.ReadAll(req.Body)
    if err != nil || len(body) == 0 {
        http.Error(res, "Invalid request", http.StatusBadRequest)
        return
    }
    originUrl := strings.TrimSpace(string(body))

    shortId := generateShortUrl(10)

    urlStorage.Lock()
    urlStorage.mapping[shortId] = originUrl
    urlStorage.Unlock()

    shortUrl := fmt.Sprintf("http://localhost:8080/%s", shortId)

    res.Header().Set("content-type", "text/plain")
    res.WriteHeader(http.StatusCreated)
    res.Write([]byte(shortUrl))
}

func retrieveUrlHandler(res http.ResponseWriter, req *http.Request) {

    if req.Method != http.MethodGet {
        http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    id := strings.TrimPrefix(req.URL.Path, "/")

    urlStorage.RLock()
    originUrl, found := urlStorage.mapping[id]
    urlStorage.RUnlock()

    if !found {
        http.Error(res, "URL not found", http.StatusBadRequest)
        return
    }

    res.Header().Set("Location", originUrl)
    res.WriteHeader(http.StatusTemporaryRedirect)
}

func main(){
    mux := http.NewServeMux()
    mux.HandleFunc("/", originUrlHandler)
    mux.HandleFunc("/{id}", retrieveUrlHandler)

    fmt.Println("Starting server on :8080...")
    err := http.ListenAndServe(":8080", mux)
    if err != nil {
        panic(err)
    }
}