// Este código implementa una API REST para gestionar partidos de fútbol utilizando Go y SQLite.
// Incluye operaciones CRUD (Crear, Leer, Actualizar, Eliminar) para los partidos y utiliza el enrutador Gorilla Mux para las rutas.
// La API permite a los clientes recuperar, crear, actualizar y eliminar partidos en una base de datos SQLite.
package main

// Importa los paquetes necesarios para la implementación de la API REST
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	_ "laligatracker/docs"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	httpSwagger "github.com/swaggo/http-swagger"
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

// MatchEvent representa un evento de un partido (gol, tarjeta amarilla o roja)
// @description Modelo que contiene la información de un evento en un partido
// @property id, team, player, minute
type MatchEvent struct {
	ID     int    `json:"id"`
	Team   string `json:"team"`
	Player string `json:"player"`
	Minute string `json:"minute"`
}

// FullMatchData representa un partido completo con eventos
// @description Modelo que contiene la información completa de un partido, incluyendo eventos
// @property id, homeTeam, awayTeam, matchDate, extraTime, homeGoals, awayGoals, goals, yellowCards, redCards
type FullMatchData struct {
	ID                   int          `json:"id"`
	HomeTeam             string       `json:"homeTeam"`
	AwayTeam             string       `json:"awayTeam"`
	MatchDate            string       `json:"matchDate"`
	ExtraTime            string       `json:"extraTime"`
	HomeGoals            int          `json:"homeGoals"`
	AwayGoals            int          `json:"awayGoals"`
	Goals                []MatchEvent `json:"goals"`
	AwayYellowCardsCount int          `json:"awayYellowCardsCount"`
	AwayRedCardsCount    int          `json:"awayRedCardsCount"`
	HomeYellowCardsCount int          `json:"homeYellowCardsCount"`
	HomeRedCardsCount    int          `json:"homeRedCardsCount"`
	YellowCards          []MatchEvent `json:"yellow_cards"`
	RedCards             []MatchEvent `json:"red_cards"`
}

// EventPayload representa la carga útil de un evento (gol, tarjeta amarilla o roja)
// @description Modelo que contiene la información de un evento en un partido
// @property Team, Player, Minute
type EventPayload struct {
	Team   string `json:"team"`
	Player string `json:"player"`
	Minute string `json:"minute"`
}

// ExtraTimePayload representa la carga útil para establecer el tiempo extra
// @description Modelo que contiene la información del tiempo extra en un partido
// @property ExtraTime
type ExtraTimePayload struct {
	ExtraTime string `json:"extraTime"`
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

		// Contar tarjetas amarillas y rojas por equipo y asignar a los campos correspondientes
		db.QueryRow("SELECT COUNT(*) FROM yellow_cards WHERE match_id = ? AND team = ?", m.ID, m.HomeTeam).Scan(&m.HomeYellowCardsCount)
		db.QueryRow("SELECT COUNT(*) FROM red_cards WHERE match_id = ? AND team = ?", m.ID, m.HomeTeam).Scan(&m.HomeRedCardsCount)
		db.QueryRow("SELECT COUNT(*) FROM yellow_cards WHERE match_id = ? AND team = ?", m.ID, m.AwayTeam).Scan(&m.AwayYellowCardsCount)
		db.QueryRow("SELECT COUNT(*) FROM red_cards WHERE match_id = ? AND team = ?", m.ID, m.AwayTeam).Scan(&m.AwayRedCardsCount)

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

	// Contar tarjetas amarillas y rojas por equipo y asignar a los campos correspondientes
	db.QueryRow("SELECT COUNT(*) FROM yellow_cards WHERE match_id = ? AND team = ?", id, m.HomeTeam).Scan(&m.HomeYellowCardsCount)
	db.QueryRow("SELECT COUNT(*) FROM yellow_cards WHERE match_id = ? AND team = ?", id, m.AwayTeam).Scan(&m.AwayYellowCardsCount)
	db.QueryRow("SELECT COUNT(*) FROM red_cards WHERE match_id = ? AND team = ?", id, m.HomeTeam).Scan(&m.HomeRedCardsCount)
	db.QueryRow("SELECT COUNT(*) FROM red_cards WHERE match_id = ? AND team = ?", id, m.AwayTeam).Scan(&m.AwayRedCardsCount)

	// Listado de goles
	m.Goals = fetchEvents("goals", id)

	// Listado de tarjetas
	m.YellowCards = fetchEvents("yellow_cards", id)
	m.RedCards = fetchEvents("red_cards", id)

	// Devolver el partido encontrado como respuesta JSON
	json.NewEncoder(w).Encode(m)
}

// fetchEvents obtiene los eventos de un partido específico
// y devuelve un slice de MatchEvent
func fetchEvents(table string, matchID string) []MatchEvent {
	// Inicializa un slice vacío para almacenar los eventos
	var events []MatchEvent

	// Ejecuta la consulta para obtener los eventos del partido específico
	// y escanea los resultados en la estructura MatchEvent
	rows, err := db.Query("SELECT id, team, player, minute FROM "+table+" WHERE match_id = ?", matchID)

	// Verifica si hubo un error al ejecutar la consulta
	// Si hubo un error, devuelve un slice vacío
	if err != nil {
		return events // Retornar slice vacío si hay error
	}
	defer rows.Close()

	// Itera sobre las filas y escanea los datos en la estructura MatchEvent
	// y agrega cada evento al slice de eventos
	for rows.Next() {
		// Crea una variable para almacenar el evento
		var e MatchEvent
		// Escanea cada fila en la estructura MatchEvent
		// y agrega el evento al slice
		if err := rows.Scan(&e.ID, &e.Team, &e.Player, &e.Minute); err == nil {
			events = append(events, e)
		}
	}
	// Retorna el slice de eventos
	return events
}

// isValidTimeFormat valida el formato de tiempo extra
// Acepta un string en formato MM:SS donde MM puede ser 0-99 y SS puede ser 00-59
func isValidTimeFormat(extraTime string) bool {
	// Valida que sea MM:SS donde MM puede ser 0–99 y SS sea 00–59
	match, _ := regexp.MatchString(`^[0-9]{1,2}:[0-5][0-9]$`, extraTime)
	return match
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

// registerEvent registra un evento (gol, tarjeta amarilla o roja) en un partido específico
// y lo inserta en la base de datos
func registerEvent(w http.ResponseWriter, r *http.Request, table string) {
	// Obtener el ID del partido de los parámetros de la URL
	id := mux.Vars(r)["id"]

	// Leer el cuerpo de la solicitud y decodificarlo en la estructura correspondiente
	var payload EventPayload

	// Verificar si hubo un error al decodificar el JSON
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Validar campos vacíos
	if payload.Team == "" || payload.Player == "" || payload.Minute == "" {
		http.Error(w, "Todos los campos son requeridos", http.StatusBadRequest)
		return
	}

	if !isValidTimeFormat(payload.Minute) {
		http.Error(w, "Formato de tiempo inválido. Usa MM:SS", http.StatusBadRequest)
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

	// Insertar el evento en la base de datos
	// Dependiendo de la tabla, se insertará en la tabla correspondiente (goals, yellow_cards o red_cards)
	_, err = db.Exec(fmt.Sprintf(`
		INSERT INTO %s (match_id, team, player, minute) 
		VALUES (?, ?, ?, ?)`, table), id, payload.Team, payload.Player, payload.Minute)

	// Verificar si hubo un error al insertar el evento
	// Si hubo un error, devolver un error 500
	// y cerrar la conexión a la base de datos
	if err != nil {
		http.Error(w, "Error al registrar el gol", http.StatusInternalServerError)
		return
	}

	// Mapeo de tabla → mensaje de respuesta
	// Dependiendo de la tabla, se asigna un mensaje diferente
	var message string
	switch table {
	case "goals":
		message = "Gol registrado correctamente"
	case "yellow_cards":
		message = "Tarjeta amarilla registrada correctamente"
	case "red_cards":
		message = "Tarjeta roja registrada correctamente"
	default:
		message = "Evento registrado correctamente"
	}

	// Devolver un mensaje de éxito como respuesta JSON
	// y cerrar la conexión a la base de datos
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

// @Summary Registrar gol
// @Description Registra un gol en un partido específico
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "ID del partido"
// @Param goal body EventPayload true "Datos del gol"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/matches/{id}/goals [patch]
func registerGoal(w http.ResponseWriter, r *http.Request) {
	registerEvent(w, r, "goals")
}

// @Summary Registrar tarjeta amarilla
// @Description Registra una tarjeta amarilla en un partido específico
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "ID del partido"
// @Param yellow_card body EventPayload true "Datos de la tarjeta amarilla"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/matches/{id}/yellow_cards [patch]
func registerYellowCard(w http.ResponseWriter, r *http.Request) {
	registerEvent(w, r, "yellow_cards")
}

// @Summary Registrar tarjeta roja
// @Description Registra una tarjeta roja en un partido específico
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "ID del partido"
// @Param red_card body EventPayload true "Datos de la tarjeta roja"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/matches/{id}/red_cards [patch]
func registerRedCard(w http.ResponseWriter, r *http.Request) {
	registerEvent(w, r, "red_cards")
}


// @Summary Establecer tiempo extra
// @Description Establece el valor de tiempo extra en un partido específico
// @Tags matches
// @Accept json
// @Produce json
// @Param id path int true "ID del partido"
// @Param extra_time body ExtraTimePayload true "Tiempo extra en formato MM:SS"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/matches/{id}/extratime [patch]
func setExtraTime(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var payload ExtraTimePayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.ExtraTime == "" {
		http.Error(w, "JSON inválido o tiempo extra faltante", http.StatusBadRequest)
		return
	}

	// Verificar que el partido exista
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM matches WHERE id=?)", id).Scan(&exists)
	if err != nil || !exists {
		http.Error(w, "Partido no encontrado", http.StatusNotFound)
		return
	}

	// Validar el formato del tiempo extra
	if !isValidTimeFormat(payload.ExtraTime) {
		http.Error(w, "Formato de tiempo inválido. Usa MM:SS", http.StatusBadRequest)
		return
	}

	// Actualizar el tiempo extra en la base de datos
	_, err = db.Exec("UPDATE matches SET extra_time=? WHERE id=?", payload.ExtraTime, id)

	// Verificar si hubo un error al actualizar el tiempo extra
	// Si hubo un error, devolver un error 500
	if err != nil {
		http.Error(w, "Error al actualizar el tiempo extra", http.StatusInternalServerError)
		return
	}

	// Devolver un mensaje de éxito como respuesta JSON
	json.NewEncoder(w).Encode(map[string]string{"message": "Tiempo extra actualizado correctamente"})
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
	r.HandleFunc("/api/matches/{id}/yellow_cards", registerYellowCard).Methods("PATCH")
	r.HandleFunc("/api/matches/{id}/red_cards", registerRedCard).Methods("PATCH")

	// Endpoint para establecer tiempo extra
	r.HandleFunc("/api/matches/{id}/extratime", setExtraTime).Methods("PATCH")

	// Endpoint para la documentación Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

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

	// Manejar solicitudes preflight (OPTIONS) para registrar tarjetas amarillas
	r.HandleFunc("/api/matches/{id}/yellow_cards", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")

	// Manejar solicitudes preflight (OPTIONS) para registrar tarjetas rojas
	r.HandleFunc("/api/matches/{id}/red_cards", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")

	// Manejar solicitudes preflight (OPTIONS) para establecer tiempo extra
	r.HandleFunc("/api/matches/{id}/extratime", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("OPTIONS")

	// Iniciar el servidor HTTP en el puerto 8080
	// y manejar las solicitudes con el enrutador configurado
	log.Println("Servidor escuchando en el puerto 8080")
	http.ListenAndServe(":8080", r)
}
