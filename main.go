
package main

import (
	"html/template"
	"net/http"
)

func main() {
	// serve static
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, nil)
	})

	println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}