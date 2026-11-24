package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

type server struct{ db *sql.DB }

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://todo:todo@localhost:5432/todo?sslmode=disable"
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	// DB mungkin belum siap saat container baru start, retry.
	for i := 0; i < 30; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		log.Printf("waiting for database: %v", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		done BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		log.Fatal(err)
	}

	s := &server{db: db}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/todos", s.list)
	mux.HandleFunc("POST /api/todos", s.create)
	mux.HandleFunc("PATCH /api/todos/{id}", s.toggle)
	mux.HandleFunc("DELETE /api/todos/{id}", s.remove)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func httpErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func (s *server) list(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(`SELECT id, title, done, created_at FROM todos ORDER BY id DESC`)
	if err != nil {
		httpErr(w, 500, err.Error())
		return
	}
	defer rows.Close()
	todos := []Todo{}
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			httpErr(w, 500, err.Error())
			return
		}
		todos = append(todos, t)
	}
	writeJSON(w, 200, todos)
}

func (s *server) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Title == "" {
		httpErr(w, 400, "title is required")
		return
	}
	var t Todo
	err := s.db.QueryRow(
		`INSERT INTO todos (title) VALUES ($1) RETURNING id, title, done, created_at`, in.Title,
	).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if err != nil {
		httpErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, t)
}

func (s *server) toggle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpErr(w, 400, "invalid id")
		return
	}
	var t Todo
	err = s.db.QueryRow(
		`UPDATE todos SET done = NOT done WHERE id = $1 RETURNING id, title, done, created_at`, id,
	).Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)
	if err == sql.ErrNoRows {
		httpErr(w, 404, "not found")
		return
	}
	if err != nil {
		httpErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, t)
}

func (s *server) remove(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		httpErr(w, 400, "invalid id")
		return
	}
	if _, err := s.db.Exec(`DELETE FROM todos WHERE id = $1`, id); err != nil {
		httpErr(w, 500, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
