package main

import (
	"fmt"
	"net/http"
	"sync"
)

var status sync.Map

func main() {
	id := 1
	for _, ip := range []string{":9200", ":9201", ":9202"} {
		sid := fmt.Sprintf("%d", id)
		status.Store(sid, http.StatusOK)
		go startServer(ip, sid)
		id++
	}
	http.HandleFunc("/down", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("down")
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method not allowed"))
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request"))
			return
		}
		_, ok := status.Load(id)
		if !ok {
			fmt.Printf("not found %s\n", id)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not found "))
			return
		}
		status.Store(id, http.StatusServiceUnavailable)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))

	})

	http.HandleFunc("/up", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("up")
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, ok := status.Load(id)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		status.Store(id, http.StatusServiceUnavailable)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("up")
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		status.Range(func(key, value interface{}) bool {
			w.Write([]byte(fmt.Sprintf("%s: %d\n", key, value)))
			return true
		})
	})

	http.ListenAndServe(":8081", nil)
}
func startServer(ip string, id string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		val, ok := status.Load(id)
		if !ok {
			fmt.Printf("not found %s\n", id)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("down"))
			return
		}
		fmt.Printf("status %s %d\n", id, val.(int))
		w.WriteHeader(val.(int))
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("hello")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello"))
	})
	http.ListenAndServe(ip, mux)
}
