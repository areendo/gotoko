package main

import (
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	ID    uint
	Name  string
	Price string
}

var db *gorm.DB
var store = sessions.NewCookieStore([]byte("secret-key"))

func main() {
	// koneksi database
	database, err := gorm.Open(sqlite.Open("gotoko.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db = database

	// auto migrate
	db.AutoMigrate(&Product{})

	// static files
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// =====================
	// CUSTOMER (PUBLIC)
	// =====================

	// halaman toko
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var products []Product
		db.Find(&products)

		tmpl := template.Must(template.ParseFiles("templates/shop.html"))
		tmpl.Execute(w, products)
	})

	// checkout
	http.HandleFunc("/checkout", func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("templates/checkout.html"))
		tmpl.Execute(w, nil)
	})

	// =====================
	// AUTH
	// =====================

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			username := r.FormValue("username")
			password := r.FormValue("password")

			if username == "admin" && password == "1234" {
				session, _ := store.Get(r, "session")
				session.Values["authenticated"] = true
				session.Save(r, w)

				http.Redirect(w, r, "/admin", http.StatusSeeOther)
				return
			}
		}

		tmpl := template.Must(template.ParseFiles("templates/login.html"))
		tmpl.Execute(w, nil)
	})

	// logout
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		session.Values["authenticated"] = false
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// =====================
	// ADMIN
	// =====================

	// dashboard
	http.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		var products []Product
		db.Find(&products)

		tmpl := template.Must(template.ParseFiles("templates/admin.html"))
		tmpl.Execute(w, products)
	})

	// tambah produk
	http.HandleFunc("/admin/add", func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if r.Method == "POST" {
			db.Create(&Product{
				Name:  r.FormValue("name"),
				Price: r.FormValue("price"),
			})

			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		tmpl := template.Must(template.ParseFiles("templates/add.html"))
		tmpl.Execute(w, nil)
	})

	// edit produk
	http.HandleFunc("/admin/edit", func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		id := r.URL.Query().Get("id")

		if r.Method == "POST" {
			db.Model(&Product{}).Where("id = ?", id).Updates(Product{
				Name:  r.FormValue("name"),
				Price: r.FormValue("price"),
			})

			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		var product Product
		db.First(&product, id)

		tmpl := template.Must(template.ParseFiles("templates/edit.html"))
		tmpl.Execute(w, product)
	})

	// delete produk
	http.HandleFunc("/admin/delete", func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		id := r.URL.Query().Get("id")
		db.Delete(&Product{}, id)

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	})

	println("Server running on http://localhost:" + getPort())
	http.ListenAndServe(":"+getPort(), nil)
}

// =====================
// HELPER
// =====================

func isAuthenticated(r *http.Request) bool {
	session, _ := store.Get(r, "session")
	auth, ok := session.Values["authenticated"].(bool)
	return ok && auth
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}