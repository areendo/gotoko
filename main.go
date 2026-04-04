package main

import (
	"html/template"
	"net/http"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	ID    uint
	Name  string
	Price string
}

var db *gorm.DB

func main() {
	database, _ := gorm.Open(sqlite.Open("gotoko.db"), &gorm.Config{})
	db = database
	db.AutoMigrate(&Product{})

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// LIST
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var products []Product
		db.Find(&products)

		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, products)
	})

	// ADD
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			db.Create(&Product{
				Name:  r.FormValue("name"),
				Price: r.FormValue("price"),
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		tmpl := template.Must(template.ParseFiles("templates/add.html"))
		tmpl.Execute(w, nil)
	})

	// DELETE
	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		db.Delete(&Product{}, id)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// EDIT
	http.HandleFunc("/edit", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		if r.Method == "POST" {
			db.Model(&Product{}).Where("id = ?", id).Updates(Product{
				Name:  r.FormValue("name"),
				Price: r.FormValue("price"),
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		var product Product
		db.First(&product, id)

		tmpl := template.Must(template.ParseFiles("templates/edit.html"))
		tmpl.Execute(w, product)
	})

	println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}