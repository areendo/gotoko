package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"gotoko/models"
)

var DB *sql.DB

func SetDB(db *sql.DB) {
	DB = db
}

func GetProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query("SELECT id,name,price,stock,category,image FROM products")
	if err != nil {
		http.Error(w, "error", 500)
		return
	}
	defer rows.Close()

	var list []models.Product

	for rows.Next() {
		var p models.Product
		rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.Category, &p.Image)
		list = append(list, p)
	}

	json.NewEncoder(w).Encode(list)
}