package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/play", playHandler)
	fmt.Println("✅ Serveur lancé : http://localhost:8088")
	http.ListenAndServe(":8088", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Bienvenue sur Drop4 Web !")
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "test")
}
