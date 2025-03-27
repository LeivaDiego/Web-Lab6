package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Match struct {
	ID        int    `json:"id"`
	HomeTeam  string `json:"homeTeam"`
	AwayTeam  string `json:"awayTeam"`
	MatchDate string `json:"matchDate"`
}

var db *sql.DB

func getMatches(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, home_team, away_team, match_date FROM matches")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var matches []Match
	for rows.Next() {
		var m Match
		rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate)
		matches = append(matches, m)
	}

	json.NewEncoder(w).Encode(matches)
}

func getMatch(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	row := db.QueryRow("SELECT id, home_team, away_team, match_date FROM matches WHERE id = ?", id)

	var m Match
	err := row.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate)
	if err != nil {
		http.Error(w, "Match not found", 404)
		return
	}

	json.NewEncoder(w).Encode(m)
}

func createMatch(w http.ResponseWriter, r *http.Request) {
	var m Match
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if m.HomeTeam == "" || m.AwayTeam == "" || m.MatchDate == "" {
		http.Error(w, "Todos los campos son obligatorios", http.StatusBadRequest)
		return
	}

	res, err := db.Exec("INSERT INTO matches (home_team, away_team, match_date) VALUES (?, ?, ?)", m.HomeTeam, m.AwayTeam, m.MatchDate)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	id, _ := res.LastInsertId()
	m.ID = int(id)
	json.NewEncoder(w).Encode(m)
}

func updateMatch(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var m Match
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if m.HomeTeam == "" || m.AwayTeam == "" || m.MatchDate == "" {
		http.Error(w, "Todos los campos son obligatorios", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE matches SET home_team=?, away_team=?, match_date=? WHERE id=?", m.HomeTeam, m.AwayTeam, m.MatchDate, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	m.ID, _ = strconv.Atoi(id)
	json.NewEncoder(w).Encode(m)
}

func deleteMatch(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	_, err := db.Exec("DELETE FROM matches WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Middleware CORS
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./database/matches.db")
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	// Agregar middleware CORS
	r.Use(enableCORS)

	// Endpoints REST
	r.HandleFunc("/api/matches", getMatches).Methods("GET")
	r.HandleFunc("/api/matches/{id}", getMatch).Methods("GET")
	r.HandleFunc("/api/matches", createMatch).Methods("POST")
	r.HandleFunc("/api/matches/{id}", updateMatch).Methods("PUT")
	r.HandleFunc("/api/matches/{id}", deleteMatch).Methods("DELETE")

	// Manejar solicitudes preflight (OPTIONS)
	r.HandleFunc("/api/matches", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")
	r.HandleFunc("/api/matches/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")

	log.Println("Servidor escuchando en el puerto 8080")
	http.ListenAndServe(":8080", r)
}
