// Este código implementa una API REST para gestionar partidos de fútbol utilizando Go y SQLite.
// Incluye operaciones CRUD (Crear, Leer, Actualizar, Eliminar) para los partidos y utiliza el enrutador Gorilla Mux para las rutas.
// La API permite a los clientes recuperar, crear, actualizar y eliminar partidos en una base de datos SQLite.
package main

// Importa los paquetes necesarios para la implementación de la API REST
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
// @property id, homeTeam, awayTeam, matchDate, extraTime
// @example { "id": 1, "homeTeam": "Real Madrid", "awayTeam": "Barcelona", "matchDate": "2025-05-10", "extraTime": "05:00" }
type Match struct {
	ID        int    `json:"id"`
	HomeTeam  string `json:"homeTeam"`
	AwayTeam  string `json:"awayTeam"`
	MatchDate string `json:"matchDate"`
	ExtraTime string `json:"extraTime"`
}

// Goal representa un gol anotado en un partido
// @description Modelo que contiene la información de un gol anotado
// @property id, team, player, minute
// @example { "id": 1, "team": "Real Madrid", "player": "Cristiano Ronaldo", "minute": "45" }
type Goal struct {
	ID     int    `json:"id"`
	Team   string `json:"team"`
	Player string `json:"player"`
	Minute string `json:"minute"`
}

// FullMatchData representa un partido completo con sus goles, tarjetas amarillas y rojas
// @description Modelo que contiene la información completa de un partido
// @property id, homeTeam, awayTeam, matchDate, extraTime, homeGoals, awayGoals
// @property goals que contiene los goles anotados en el partido
// @example { "id": 1, "homeTeam": "Real Madrid", "awayTeam": "Barcelona", "matchDate": "2025-05-10", "extraTime": "05:00", "homeGoals": 2, "awayGoals": 1, "goals": [{ "id": 1, "team": "Real Madrid", "player": "Cristiano Ronaldo", "minute": "45" }] }
type FullMatchData struct {
	ID         int    `json:"id"`
	HomeTeam   string `json:"homeTeam"`
	AwayTeam   string `json:"awayTeam"`
	MatchDate  string `json:"matchDate"`
	ExtraTime  string `json:"extraTime"`
	HomeGoals  int    `json:"homeGoals"`
	AwayGoals  int    `json:"awayGoals"`
	Goals      []Goal `json:"goals,omitempty"` // solo se usa en getMatch
}

// db es la variable global que representa la conexión a la base de datos SQLite
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
	rows, err := db.Query("SELECT id, home_team, away_team, match_date, extra_time FROM matches")	
	
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
	var matches []FullMatchData

	// Iterar sobre las filas y escanear los datos en la estructura Match
	for rows.Next() {
		// Crear una variable para almacenar el partido
		var m FullMatchData

		// Escanear cada fila en la estructura Match y agregarla al slice
		err := rows.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate, &m.ExtraTime)

		// Verificar si hubo un error al escanear la fila
		// Si hubo un error, devolver un error 500
		// y cerrar la conexión a la base de datos
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Contar goles por equipo y asignar a los campos correspondientes
		db.QueryRow("SELECT COUNT(*) FROM goals WHERE match_id = ? AND team = ?", m.ID, m.HomeTeam).Scan(&m.HomeGoals)
		db.QueryRow("SELECT COUNT(*) FROM goals WHERE match_id = ? AND team = ?", m.ID, m.AwayTeam).Scan(&m.AwayGoals)


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
	row := db.QueryRow("SELECT id, home_team, away_team, match_date, extra_time FROM matches WHERE id = ?", id)

	// Crear una variable para almacenar el partido
	var m FullMatchData

	// Escanear la fila en la estructura Match
	err := row.Scan(&m.ID, &m.HomeTeam, &m.AwayTeam, &m.MatchDate, &m.ExtraTime)
	
	// Verificar si hubo un error al escanear la fila
	// Si hubo un error, devolver un error 404
	// y cerrar la conexión a la base de datos
	if err != nil {
		http.Error(w, "Partido no encontrado", 404)
		return
	}

	// Contar goles por equipo y asignar a los campos correspondientes
	db.QueryRow("SELECT COUNT(*) FROM goals WHERE match_id = ? AND team = ?", id, m.HomeTeam).Scan(&m.HomeGoals)
	db.QueryRow("SELECT COUNT(*) FROM goals WHERE match_id = ? AND team = ?", id, m.AwayTeam).Scan(&m.AwayGoals)

	// Obtener detalles de goles del partido
	// Ejecutar la consulta para obtener los goles del partido
	golRows, err := db.Query("SELECT id, team, player, minute FROM goals WHERE match_id = ?", id)
	
	// Verificar si hubo un error al ejecutar la consulta
	// Si hubo un error, devolver un error 500
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	// Asegurarse de cerrar las filas después de usarlas
	defer golRows.Close()

	// Crear un slice para almacenar los goles
	var goles []Goal
	// Iterar sobre las filas y escanear los datos en la estructura Goal
	for golRows.Next() {
		// Crear una variable para almacenar el gol
		var g Goal
		// Escanear cada fila en la estructura Goal y agregarla al slice
		err := golRows.Scan(&g.ID, &g.Team, &g.Player, &g.Minute)
		// Verificar si hubo un error al escanear la fila
		// Si hubo un error, devolver un error 500
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// Agregar el gol al slice
		goles = append(goles, g)
	}
	// Asignar los goles al campo correspondiente de la estructura Match
	m.Goals = goles

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
	res, err := db.Exec(`INSERT INTO matches (home_team, away_team, match_date) VALUES (?, ?, ?)`, m.HomeTeam, m.AwayTeam, m.MatchDate)
	
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


// @Summary Registrar un gol
// @Description Registra un gol con el nombre del jugador, equipo y minuto
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "ID del partido"
// @Param body body map[string]string true "Datos del gol"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/matches/{id}/goals [patch]
func registerGoal(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var payload struct {
		Team   string `json:"team"`
		Player string `json:"player"`
		Minute string `json:"minute"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Validar campos vacíos
	if payload.Team == "" || payload.Player == "" || payload.Minute == "" {
		http.Error(w, "Todos los campos son requeridos", http.StatusBadRequest)
		return
	}

	// Verificar si el partido existe y obtener nombres reales de los equipos
	var home, away string
	err := db.QueryRow("SELECT home_team, away_team FROM matches WHERE id = ?", id).Scan(&home, &away)
	if err != nil {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}

	// Validar que el equipo exista en este partido
	if payload.Team != home && payload.Team != away {
		http.Error(w, "El equipo no corresponde al partido", http.StatusBadRequest)
		return
	}

	// Insertar el gol en la tabla goals
	_, err = db.Exec(`
		INSERT INTO goals (match_id, team, player, minute)
		VALUES (?, ?, ?, ?)`, id, payload.Team, payload.Player, payload.Minute)
	if err != nil {
		http.Error(w, "Error al registrar el gol", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Gol registrado correctamente"})
}



// enableCORS configura los encabezados necesarios para permitir solicitudes desde otros orígenes (CORS)
// Se aplica como middleware para todas las rutas.
func enableCORS(next http.Handler) http.Handler {
	
	// Configura los encabezados CORS para permitir solicitudes desde cualquier origen
	// y permite métodos y encabezados específicos
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
		// Configura los encabezados CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
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
	
	// Enpoints PATCH para registrar goles, tarjetas amarillas y rojas
	r.HandleFunc("/api/matches/{id}/goals", registerGoal).Methods("PATCH")

	// Manejar solicitudes preflight (OPTIONS)
	r.HandleFunc("/api/matches", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")
	
	r.HandleFunc("/api/matches/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")

	// Manejar solicitudes preflight (OPTIONS) para registrar goles
	r.HandleFunc("/api/matches/{id}/goals", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	  }).Methods("OPTIONS")
	  

	// Iniciar el servidor HTTP en el puerto 8080
	// y manejar las solicitudes con el enrutador configurado
	log.Println("Servidor escuchando en el puerto 8080")
	http.ListenAndServe(":8080", r)
}