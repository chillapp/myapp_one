package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

const appVersion = "0.5.2"

var config = map[string]string{
	"DB":       os.Getenv("DB"),
	"DB_URI":   os.Getenv("DB_URI"),
	"HOSTNAME": os.Getenv("HOSTNAME"),
	"GREETING": os.Getenv("GREETING"),
}

func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/version", version)
	http.HandleFunc("/create", insertRow)
	http.HandleFunc("/config", getConfig)
	// Слушаем путь localhost:8080/health
	http.HandleFunc("/health", health)
	// Стартуем web-сервер
	log.Println("web-server running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

// Endpoint для приветствия
func hello(w http.ResponseWriter, r *http.Request) {
	response := fmt.Sprintf("Hello world from %s\n", config["HOSTNAME"])
	_, _ = w.Write([]byte(response))
}

// Endpoint для получения версии
func version(w http.ResponseWriter, r *http.Request) {
	response := fmt.Sprintf("{\"version\":\"%s\"}\n", appVersion)
	_, _ = w.Write([]byte(response))
}

// Endpoint для тестирования сервиса
func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 page not found"))
		return
	}
	response := "{\"status\":\"ok\"}\n"
	_, _ = w.Write([]byte(response))
}

// Endpoint для вставки строки в БД
func insertRow(w http.ResponseWriter, r *http.Request) {
	name, ok := r.URL.Query()["name"]
	if !ok || name[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
		return
	}
	db, err := sql.Open("mysql", config["DB_URI"]+"/")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config["DB"])
	_, err = db.Exec(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	_ = db.Close()
	dbUri := fmt.Sprintf("%s/%s", config["DB_URI"], config["DB"])
	db, err = sql.Open("mysql", dbUri)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	defer db.Close()

	query = `
		CREATE TABLE IF NOT EXISTS users (
			ID INT AUTO_INCREMENT PRIMARY KEY,
			NAME VARCHAR(64) UNIQUE
		)
	`
	_, err = db.Exec(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	query = `INSERT INTO users SET name = ?`
	res, err := db.Exec(query, name[0])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	response := fmt.Sprintf("{\"id\":%d\"name\":\"%s\"}\n", id, name[0])
	_, _ = w.Write([]byte(response))
}

// Endpoint для получения конфига
func getConfig(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(config)
	_, _ = w.Write(data)
}
