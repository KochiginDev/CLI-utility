package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
)

type Ram_Storage struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type Storage struct {
	data map[string]Ram_Storage
	mu   sync.RWMutex
}
type Object interface {
	Get(string) (string, error)
	Post(string, string) (string, error)
	Put(string, string) (string, error)
	Delete(string) (string, error)
}
type Router interface {
	Method(*http.Request) Object //Вызывает определенный handler
}

func (s *Storage) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Path[1:]
	if key == "" {
		http.Error(w, "Missing key", http.StatusBadRequest)
		return
	}
	item, exists := s.data[key]
	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	err := json.NewEncoder(w).Encode(item)
	if err != nil {
		return
	}
	return
}
func (s *Storage) PostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Path[1:]
	if key == "" {
		http.Error(w, "Key is required in URL path", http.StatusBadRequest)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	return
}

func IsValidURL(u string) bool {
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return false
	}
	return true
}
func main() {
	storage := &Storage{data: make(map[string]Ram_Storage),
		mu: sync.RWMutex{}}
	http.HandleFunc("/get", storage.GetHandler)
	http.HandleFunc("/post", storage.PostHandler)
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		panic(err)
	}
}
