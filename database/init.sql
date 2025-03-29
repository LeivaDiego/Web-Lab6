-- ================================================================
-- Script para inicializar la base de datos de La Liga Tracker
-- Crea la tabla de partidos, goles, tarjetas amarillas y rojas
-- ================================================================

-- Tabla de partidos
CREATE TABLE IF NOT EXISTS matches (
  id INTEGER PRIMARY KEY AUTOINCREMENT,               -- ID del partido
  home_team TEXT NOT NULL,                            -- Nombre del equipo local
  away_team TEXT NOT NULL,                            -- Nombre del equipo visitante
  match_date TEXT NOT NULL,                           -- Fecha del partido (YYYY-MM-DD)
  extra_time TEXT DEFAULT '00:00'                     -- Tiempo extra en formato MM:SS
);

-- Tabla de goles
CREATE TABLE IF NOT EXISTS goals (
  id INTEGER PRIMARY KEY AUTOINCREMENT,               -- ID del gol
  match_id INTEGER NOT NULL,                          -- Referencia al partido
  team TEXT NOT NULL,                                 -- Nombre del equipo que anotó
  player TEXT NOT NULL,                               -- Jugador que anotó
  minute TEXT NOT NULL,                               -- Minuto del gol (MM:SS)
  FOREIGN KEY (match_id) REFERENCES matches(id)       -- Relación con la tabla de partidos
);

-- Tabla de tarjetas amarillas
CREATE TABLE IF NOT EXISTS yellow_cards (
  id INTEGER PRIMARY KEY AUTOINCREMENT,               -- ID de la tarjeta amarilla
  match_id INTEGER NOT NULL,                          -- Referencia al partido
  team TEXT NOT NULL,                                 -- Nombre del equipo que recibió la tarjeta
  player TEXT NOT NULL,                               -- Jugador que recibió la tarjeta
  minute TEXT NOT NULL,                               -- Minuto de la tarjeta (MM:SS)
  FOREIGN KEY (match_id) REFERENCES matches(id)       -- Relación con la tabla de partidos
);

-- Tabla de tarjetas rojas
CREATE TABLE IF NOT EXISTS red_cards (
  id INTEGER PRIMARY KEY AUTOINCREMENT,               -- ID de la tarjeta roja
  match_id INTEGER NOT NULL,                          -- Referencia al partido
  team TEXT NOT NULL,                                 -- Nombre del equipo que recibió la tarjeta
  player TEXT NOT NULL,                               -- Jugador que recibió la tarjeta
  minute TEXT NOT NULL,                               -- Minuto de la tarjeta (MM:SS)
  FOREIGN KEY (match_id) REFERENCES matches(id)       -- Relación con la tabla de partidos
);

-- ===============================
-- DATOS DE EJEMPLO
-- ===============================

-- Partidos
INSERT INTO matches (home_team, away_team, match_date, extra_time) VALUES
  ('Real Madrid', 'Barcelona', '2025-05-10', '05:00'),
  ('Atletico Madrid', 'Valencia', '2025-06-01', '02:30'),
  ('Sevilla', 'Villarreal', '2025-06-15', '00:00'),
  ('Boca Juniors', 'River Plate', '2025-07-20', '07:45');

-- Goles
INSERT INTO goals (match_id, team, player, minute) VALUES
  (1, 'Real Madrid', 'Vinicius Jr.', '12:34'),
  (1, 'Barcelona', 'Lewandowski', '21:12'),
  (2, 'Atletico Madrid', 'Griezmann', '44:00'),
  (4, 'River Plate', 'Borja', '05:55');

-- Tarjetas Amarillas
INSERT INTO yellow_cards (match_id, team, player, minute) VALUES
  (1, 'Real Madrid', 'Carvajal', '35:00'),
  (1, 'Barcelona', 'Gavi', '36:20'),
  (2, 'Valencia', 'Paulista', '60:00');

-- Tarjetas Rojas
INSERT INTO red_cards (match_id, team, player, minute) VALUES
  (2, 'Valencia', 'Paulista', '88:00'),
  (4, 'Boca Juniors', 'Rojo', '70:00');
