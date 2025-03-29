// Propósito: Implementar una API REST para gestionar partidos de fútbol usando Go y SQLite
// Descripción: Este código implementa una API REST para gestionar partidos de fútbol utilizando Go y SQLite.
// Incluye operaciones CRUD (Crear, Leer, Actualizar, Eliminar) para los partidos y utiliza el enrutador Gorilla Mux para las rutas.
// La API permite a los clientes recuperar, crear, actualizar y eliminar partidos en una base de datos SQLite.
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


// Match representa un partido de fútbol
// @description Modelo que contiene la información básica de un partido
// @property id, homeTeam, awayTeam, matchDate, goals, yellowCards, redCards, extraTime
// @example { "id": 1, "homeTeam": "Real Madrid", "awayTeam": "Barcelona", "matchDate": "2025-05-10", "goals": 3, "yellowCards": 1, "redCards": 0, "extraTime": "05:00" }
type Match struct {
	ID          int    `json:"id"`
	HomeTeam    string `json:"homeTeam"`
	AwayTeam    string `json:"awayTeam"`
	MatchDate   string `json:"matchDate"`
	Goals       int    `json:"goals"`
	YellowCards int    `json:"yellowCards"`
	RedCards    int    `json:"redCards"`
	ExtraTime   string `json:"extraTime"`
}


var db *sql.DB


// @Summary Obtener todos los partidos
// @Description Retorna una lista con todos los partidos registrados
// @Tags matches
// @Accept json
// @Produce json
// @Success 200 {array} Match
// @Router /api/matches [get]
func getMatches(w http.ResponseWriter, r *http.Request) {
	// Ejecutar la consulta para obtener todos los partidos
	rows, err := db.Query("SELECT id, home_team, away_team, match_date, goals, yellow_cards, red_cards, extra_time FROM matches")
	
	// Verificar si hubo un error al ejecutar la consulta
	// Si hubo un error, devolver un error 500
	// y cerrar la conexión a la base de datos
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Asegurarse de cerrar las filas después de usarlas
	defer rows.Close()

	// Crear un slice para almacenar los partidos
	var matches []Match

	// Iterar sobre las filas y escanear los datos en la estructura Match
	for rows.Next() {
		// Crear una variable para almacenar el partido
		var m Match

		// Escanear cada fila en la estructura Match y agregarla al slice
		err := rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate, &m.Goals, &m.YellowCards, &m.RedCards, &m.ExtraTime)
		
		// Verificar si hubo un error al escanear la fila
		// Si hubo un error, devolver un error 500
		// y cerrar la conexión a la base de datos
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Agregar el partido al slice
		matches = append(matches, m)
	}

	// Verificar si hubo un error al iterar sobre las filas
	json.NewEncoder(w).Encode(matches)
}


// @Summary Obtener partido por ID
// @Description Retorna los datos de un partido específico
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "ID del partido"
// @Success 200 {object} Match
// @Failure 404 {object} map[string]string
// @Router /api/matches/{id} [get]
func getMatch(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del partido de los parámetros de la URL
	id := mux.Vars(r)["id"]
	// Ejecutar la consulta para obtener el partido por ID
	row := db.QueryRow("SELECT id, home_team, away_team, match_date, goals, yellow_cards, red_cards, extra_time FROM matches WHERE id = ?", id)

	// Crear una variable para almacenar el partido
	var m Match

	// Escanear la fila en la estructura Match
	err := row.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate, &m.Goals, &m.YellowCards, &m.RedCards, &m.ExtraTime)
	
	// Verificar si hubo un error al escanear la fila
	// Si hubo un error, devolver un error 404
	// y cerrar la conexión a la base de datos
	if err != nil {
		http.Error(w, "Partido no encontrado", 404)
		return
	}

	// Devolver el partido encontrado como respuesta JSON
	json.NewEncoder(w).Encode(m)
}


// @Summary Crear un nuevo partido
// @Description Crea un nuevo registro de partido con los datos básicos
// @Tags matches
// @Accept json
// @Produce json
// @Param match body Match true "Datos del partido"
// @Success 200 {object} Match
// @Failure 400 {object} map[string]string
// @Router /api/matches [post]
func createMatch(w http.ResponseWriter, r *http.Request) {
	// Leer el cuerpo de la solicitud y decodificarlo en la estructura Match
	var m Match
	err := json.NewDecoder(r.Body).Decode(&m)
	// Verificar si hubo un error al decodificar el JSON
	// Si hubo un error, devolver un error 400
	// y cerrar la conexión a la base de datos
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Verificar si los campos requeridos están presentes
	// Si faltan campos, devolver un error 400
	// y cerrar la conexión a la base de datos
	// Los campos requeridos son homeTeam, awayTeam y matchDate
	if m.HomeTeam == "" || m.AwayTeam == "" || m.MatchDate == "" {
		http.Error(w, "Todos los campos son obligatorios", http.StatusBadRequest)
		return
	}

	// InsertaR solo los campos requeridos, los demás se usarán los valores por defecto
	res, err := db.Exec(`INSERT INTO matches (home_team, away_team, match_date) VALUES (?, ?, ?)`,
		m.HomeTeam, m.AwayTeam, m.MatchDate)
	
	// Verificar si hubo un error al insertar el partido
	// Si hubo un error, devolver un error 500
	// y cerrar la conexión a la base de datos
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Obtener el ID del nuevo partido insertado
	// y asignar valores por defecto a los demás campos
	id, _ := res.LastInsertId()
	m.ID = int(id)
	m.Goals = 0
	m.YellowCards = 0
	m.RedCards = 0
	m.ExtraTime = "00:00"
	json.NewEncoder(w).Encode(m)
}

// @Summary Actualizar partido
// @Description Modifica los datos de un partido existente por ID
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "ID del partido"
// @Param match body Match true "Datos actualizados"
// @Success 200 {object} Match
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/matches/{id} [put]
func updateMatch(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del partido de los parámetros de la URL
	// y leer el cuerpo de la solicitud para decodificarlo en la estructura Match
	id := mux.Vars(r)["id"]
	var m Match
	err := json.NewDecoder(r.Body).Decode(&m)
	
	// Verificar si hubo un error al decodificar el JSON
	// Si hubo un error, devolver un error 400
	// y cerrar la conexión a la base de datos
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Verificar si los campos requeridos están presentes
	// Si faltan campos, devolver un error 400
	// y cerrar la conexión a la base de datos
	// Los campos requeridos son homeTeam, awayTeam y matchDate
	if m.HomeTeam == "" || m.AwayTeam == "" || m.MatchDate == "" {
		http.Error(w, "Todos los campos son obligatorios", http.StatusBadRequest)
		return
	}

	// Solo actualizar los campos requeridos, los opcionales se mantienen sin cambios (para eso se usará PATCH) 
	_, err = db.Exec(`UPDATE matches SET home_team=?, away_team=?, match_date=? WHERE id=?`,
		m.HomeTeam, m.AwayTeam, m.MatchDate, id)
	
	// Verificar si hubo un error al actualizar el partido
	// Si hubo un error, devolver un error 500
	// y cerrar la conexión a la base de datos
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Asignar el ID del partido actualizado a la estructura Match
	// y devolver el partido actualizado como respuesta JSON
	m.ID, _ = strconv.Atoi(id)
	json.NewEncoder(w).Encode(m)
}

// @Summary Eliminar partido
// @Description Elimina un partido de la base de datos por ID
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "ID del partido"
// @Success 204 {string} string "Sin contenido"
// @Failure 500 {object} map[string]string
// @Router /api/matches/{id} [delete]
func deleteMatch(w http.ResponseWriter, r *http.Request) {
	// Obtener el ID del partido de los parámetros de la URL
	// y ejecutar la consulta para eliminar el partido por ID
	id := mux.Vars(r)["id"]
	_, err := db.Exec("DELETE FROM matches WHERE id=?", id)
	
	// Verificar si hubo un error al eliminar el partido
	// Si hubo un error, devolver un error 500
	// y cerrar la conexión a la base de datos
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Devolver un código de estado 204 (Sin contenido) si la eliminación fue exitosa
	// y cerrar la conexión a la base de datos
	w.WriteHeader(http.StatusNoContent)
}

// enableCORS configura los encabezados necesarios para permitir solicitudes desde otros orígenes (CORS)
// Se aplica como middleware para todas las rutas.
func enableCORS(next http.Handler) http.Handler {
	
	// Configura los encabezados CORS para permitir solicitudes desde cualquier origen
	// y permite métodos y encabezados específicos
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
		// Configura los encabezados CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Maneja las solicitudes preflight (OPTIONS) para permitir el intercambio de recursos entre orígenes
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Llama al siguiente manejador en la cadena de middleware
		// para procesar la solicitud real
		next.ServeHTTP(w, r)
	})
}

// main inicializa la conexión a la base de datos, configura las rutas y arranca el servidor HTTP
func main() {
	// Inicializa la conexión a la base de datos SQLite
	var err error
	db, err = sql.Open("sqlite3", "./database/matches.db")
	
	// Verifica si hubo un error al abrir la base de datos
	// Si hubo un error, imprime el error y termina la ejecución del programa
	// y cierra la conexión a la base de datos
	if err != nil {
		log.Fatal(err)
	}

	// Verifica si la base de datos está accesible
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

	// Iniciar el servidor HTTP en el puerto 8080
	// y manejar las solicitudes con el enrutador configurado
	log.Println("Servidor escuchando en el puerto 8080")
	http.ListenAndServe(":8080", r)
}