-- Este script crea una base de datos SQLite para 
-- almacenar información sobre partidos de fútbol, incluyendo equipos, jugadores y estadísticas.


-- Crea la base de datos y las tablas necesarias para almacenar la información.
CREATE TABLE IF NOT EXISTS matches (
  id INTEGER PRIMARY KEY AUTOINCREMENT,     -- ID del partido (clave primaria)
  home_team TEXT NOT NULL,                  -- Nombre del equipo local
  away_team TEXT NOT NULL,                  -- Nombre del equipo visitante
  match_date TEXT NOT NULL,                 -- Fecha del partido en formato DD-MM-YYYY
  goals INTEGER DEFAULT 0,                  -- Cantidad de goles anotados
  yellow_cards INTEGER DEFAULT 0,           -- Cantidad de tarjetas amarillas
  red_cards INTEGER DEFAULT 0,              -- Cantidad de tarjetas  rojas
  extra_time TEXT DEFAULT '00:00'           -- Tiempo extra en formato MM:SS
);


-- Inserta datos de ejemplo en la tabla de partidos.
-- Estos datos son ficticios y se utilizan solo para propósitos de demostración.
INSERT INTO matches (home_team, away_team, match_date, goals, yellow_cards, red_cards, extra_time) VALUES
  ('Real Madrid', 'Barcelona', '2025-05-10', 3, 2, 0, '05:00'),
  ('Atletico Madrid', 'Valencia', '2025-06-01', 1, 1, 1, '02:30'),
  ('Sevilla', 'Villarreal', '2025-06-15', 0, 0, 0, '00:00'),
  ('Boca Juniors', 'River Plate', '2025-07-20', 2, 3, 1, '07:45');
