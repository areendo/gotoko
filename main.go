package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

var adminUser = "admin"
var adminPass = "12345"
var adminWA = "6281234567890"

func main() {
	db, _ = sql.Open("sqlite", "./toko.db")
	os.Mkdir("uploads", os.ModePerm)

	db.Exec(`CREATE TABLE IF NOT EXISTS products(
	id INTEGER PRIMARY KEY,
	name TEXT,
	price INTEGER,
	stock INTEGER,
	category TEXT,
	image TEXT)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS orders(
	id INTEGER PRIMARY KEY,
	name TEXT,
	items TEXT,
	total INTEGER,
	status TEXT,
	time TEXT)`)

	http.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	http.HandleFunc("/api/products", getProducts)
	http.HandleFunc("/api/create", createProduct)
	http.HandleFunc("/api/delete", deleteProduct)
	http.HandleFunc("/api/update-product", updateProduct)
	http.HandleFunc("/api/categories", getCategories)
	http.HandleFunc("/api/upload", uploadImage)

	http.HandleFunc("/api/orders", getOrders)
	http.HandleFunc("/api/update", updateOrder)
	http.HandleFunc("/api/checkout", checkout)

	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)

	fmt.Println("jalan di http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func isAuth(r *http.Request) bool {
	c, err := r.Cookie("session")
	if err != nil {
		return false
	}
	return c.Value == "ok"
}

func getProducts(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.Query("SELECT id,name,price,stock,category,image FROM products")
	var list []map[string]interface{}

	for rows.Next() {
		var id, price, stock int
		var name, cat, img string
		rows.Scan(&id, &name, &price, &stock, &cat, &img)

		list = append(list, map[string]interface{}{
			"id": id, "name": name, "price": price,
			"stock": stock, "category": cat, "image": img,
		})
	}
	json.NewEncoder(w).Encode(list)
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	if !isAuth(r) {
		http.Error(w, "unauthorized", 401)
		return
	}

	price, _ := strconv.Atoi(r.FormValue("price"))
	stock, _ := strconv.Atoi(r.FormValue("stock"))

	db.Exec("INSERT INTO products(name,price,stock,category,image) VALUES(?,?,?,?,?)",
		r.FormValue("name"),
		price,
		stock,
		r.FormValue("category"),
		r.FormValue("image"),
	)
}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	if !isAuth(r) {
		http.Error(w, "unauthorized", 401)
		return
	}

	price, _ := strconv.Atoi(r.FormValue("price"))
	stock, _ := strconv.Atoi(r.FormValue("stock"))

	db.Exec("UPDATE products SET name=?,price=?,stock=?,category=?,image=? WHERE id=?",
		r.FormValue("name"),
		price,
		stock,
		r.FormValue("category"),
		r.FormValue("image"),
		r.FormValue("id"),
	)
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	db.Exec("DELETE FROM products WHERE id=?", r.FormValue("id"))
}

func getCategories(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.Query("SELECT DISTINCT category FROM products")
	var list []string
	for rows.Next() {
		var c string
		rows.Scan(&c)
		list = append(list, c)
	}
	json.NewEncoder(w).Encode(list)
}

func uploadImage(w http.ResponseWriter, r *http.Request) {
	if !isAuth(r) {
		http.Error(w, "unauthorized", 401)
		return
	}

	r.ParseMultipartForm(10 << 20)

	file, handler, _ := r.FormFile("image")
	defer file.Close()

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)

	dst, _ := os.Create("./uploads/" + filename)
	defer dst.Close()

	io.Copy(dst, file)

	json.NewEncoder(w).Encode(map[string]string{
		"url": "/uploads/" + filename,
	})
}

func getOrders(w http.ResponseWriter, r *http.Request) {
	rows, _ := db.Query("SELECT id,name,items,total,status,time FROM orders")

	var list []map[string]interface{}

	for rows.Next() {
		var id, total int
		var name, items, status, time string

		rows.Scan(&id, &name, &items, &total, &status, &time)

		list = append(list, map[string]interface{}{
			"id": id, "name": name, "items": items,
			"total": total, "status": status, "time": time,
		})
	}

	json.NewEncoder(w).Encode(list)
}

func updateOrder(w http.ResponseWriter, r *http.Request) {
	db.Exec("UPDATE orders SET status=? WHERE id=?", r.FormValue("status"), r.FormValue("id"))
}

func checkout(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Name  string
		Items string
		Total int
	}

	json.NewDecoder(r.Body).Decode(&d)

	now := time.Now().Format("2006-01-02 15:04")

	db.Exec("INSERT INTO orders(name,items,total,status,time) VALUES(?,?,?,?,?)",
		d.Name, d.Items, d.Total, "pending", now)

	msg := fmt.Sprintf("Nama:%s\n%s\nTotal:%d",
		d.Name, strings.ReplaceAll(d.Items, ", ", "\n"), d.Total)

	wa := "https://wa.me/" + adminWA + "?text=" + url.QueryEscape(msg)

	json.NewEncoder(w).Encode(map[string]string{"wa": wa})
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("username") == adminUser && r.FormValue("password") == adminPass {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "ok", Path: "/"})
		w.Write([]byte("ok"))
		return
	}
	http.Error(w, "fail", 401)
}

func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", MaxAge: -1, Path: "/"})
}