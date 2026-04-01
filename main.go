package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"gotoko/handlers"

	_ "modernc.org/sqlite"
)

func main() {
	db, _ := sql.Open("sqlite", "./toko.db")
	os.Mkdir("uploads", os.ModePerm)

	handlers.SetDB(db)

	http.Handle("/", http.FileServer(http.Dir("./templates")))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	http.HandleFunc("/api/products", handlers.GetProducts)

	fmt.Println("jalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}