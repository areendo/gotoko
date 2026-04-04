package main

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	ID          uint `gorm:"primaryKey"`
	Name        string
	Price       int
	Description string
}

type CartPageData struct {
	Products   []Product
	Total      int
	CartCount  int
	GrandTotal int
}

var db *gorm.DB
var store = sessions.NewCookieStore([]byte("super-secret-key"))

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("gotoko.db"), &gorm.Config{})
	if err != nil {
		panic("gagal konek ke database")
	}

	err = db.AutoMigrate(&Product{})
	if err != nil {
		panic("gagal migrate database")
	}

	seedProducts()

	http.HandleFunc("/", shopHandler)
	http.HandleFunc("/cart", cartHandler)
	http.HandleFunc("/cart/add", cartAddHandler)
	http.HandleFunc("/cart/remove", cartRemoveHandler)
	http.HandleFunc("/checkout", checkoutHandler)

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)

	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/admin/add", adminAddHandler)
	http.HandleFunc("/admin/edit", adminEditHandler)
	http.HandleFunc("/admin/delete", adminDeleteHandler)

	println("Server running on http://localhost:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func seedProducts() {
	var count int64
	db.Model(&Product{}).Count(&count)
	if count == 0 {
		products := []Product{
			{Name: "Kopi Susu", Price: 18000, Description: "Minuman kopi susu favorit pelanggan."},
			{Name: "Roti Bakar Coklat", Price: 15000, Description: "Roti bakar lembut dengan topping coklat."},
			{Name: "Mie Goreng Spesial", Price: 25000, Description: "Mie goreng dengan topping lengkap."},
		}
		db.Create(&products)
	}
}

func renderTemplate(w http.ResponseWriter, file string, data any) {
	tmpl, err := template.ParseFiles("templates/" + file)
	if err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func getSession(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, "gotoko-session")
}

func getCartIDs(r *http.Request) []int {
	session, err := getSession(r)
	if err != nil {
		return []int{}
	}

	raw, ok := session.Values["cart"]
	if !ok {
		return []int{}
	}

	cartIDs, ok := raw.([]int)
	if !ok {
		return []int{}
	}

	return cartIDs
}

func saveCartIDs(w http.ResponseWriter, r *http.Request, cartIDs []int) {
	session, err := getSession(r)
	if err != nil {
		return
	}
	session.Values["cart"] = cartIDs
	_ = session.Save(r, w)
}

func getCartProducts(r *http.Request) ([]Product, int) {
	cartIDs := getCartIDs(r)
	products := []Product{}
	total := 0

	for _, id := range cartIDs {
		var product Product
		if err := db.First(&product, id).Error; err == nil {
			products = append(products, product)
			total += product.Price
		}
	}

	return products, total
}

func getCartCount(r *http.Request) int {
	return len(getCartIDs(r))
}

func isAuthenticated(r *http.Request) bool {
	session, err := getSession(r)
	if err != nil {
		return false
	}

	auth, ok := session.Values["authenticated"].(bool)
	return ok && auth
}

func requireAuth(w http.ResponseWriter, r *http.Request) bool {
	if !isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return false
	}
	return true
}

func shopHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var products []Product
	db.Order("id desc").Find(&products)

	data := struct {
		Products  []Product
		CartCount int
	}{
		Products:  products,
		CartCount: getCartCount(r),
	}

	renderTemplate(w, "shop.html", data)
}

func cartAddHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var product Product
	if err := db.First(&product, id).Error; err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	cartIDs := getCartIDs(r)
	cartIDs = append(cartIDs, id)
	saveCartIDs(w, r, cartIDs)

	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func cartHandler(w http.ResponseWriter, r *http.Request) {
	products, total := getCartProducts(r)

	data := CartPageData{
		Products:  products,
		Total:     total,
		CartCount: len(products),
	}

	renderTemplate(w, "cart.html", data)
}

func cartRemoveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	indexStr := r.FormValue("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	cartIDs := getCartIDs(r)
	if index < 0 || index >= len(cartIDs) {
		http.Redirect(w, r, "/cart", http.StatusSeeOther)
		return
	}

	cartIDs = append(cartIDs[:index], cartIDs[index+1:]...)
	saveCartIDs(w, r, cartIDs)

	http.Redirect(w, r, "/cart", http.StatusSeeOther)
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	products, total := getCartProducts(r)

	if r.Method == http.MethodPost {
		saveCartIDs(w, r, []int{})
		data := struct {
			Success bool
			Total   int
		}{
			Success: true,
			Total:   total,
		}
		renderTemplate(w, "checkout.html", data)
		return
	}

	data := CartPageData{
		Products:   products,
		GrandTotal: total,
		CartCount:  len(products),
	}

	renderTemplate(w, "checkout.html", data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	type LoginData struct {
		Error string
	}

	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "admin" && password == "1234" {
			session, _ := getSession(r)
			session.Values["authenticated"] = true
			_ = session.Save(r, w)
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		renderTemplate(w, "login.html", LoginData{
			Error: "Username atau password salah",
		})
		return
	}

	renderTemplate(w, "login.html", LoginData{})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := getSession(r)
	session.Values["authenticated"] = false
	_ = session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(w, r) {
		return
	}

	var products []Product
	db.Order("id desc").Find(&products)

	data := struct {
		Products []Product
	}{
		Products: products,
	}

	renderTemplate(w, "admin.html", data)
}

func adminAddHandler(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(w, r) {
		return
	}

	type FormData struct {
		Error string
	}

	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		priceStr := r.FormValue("price")
		description := r.FormValue("description")

		price, err := strconv.Atoi(priceStr)
		if err != nil {
			renderTemplate(w, "add.html", FormData{
				Error: "Harga harus angka",
			})
			return
		}

		product := Product{
			Name:        name,
			Price:       price,
			Description: description,
		}
		db.Create(&product)

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "add.html", FormData{})
}

func adminEditHandler(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(w, r) {
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	var product Product
	if err := db.First(&product, id).Error; err != nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	type EditData struct {
		Product Product
		Error   string
	}

	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		priceStr := r.FormValue("price")
		description := r.FormValue("description")

		price, err := strconv.Atoi(priceStr)
		if err != nil {
			renderTemplate(w, "edit.html", EditData{
				Product: product,
				Error:   "Harga harus angka",
			})
			return
		}

		product.Name = name
		product.Price = price
		product.Description = description
		db.Save(&product)

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	renderTemplate(w, "edit.html", EditData{
		Product: product,
	})
}

func adminDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !requireAuth(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err == nil {
		db.Delete(&Product{}, id)
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}