package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	db *pgxpool.Pool
}

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Consumption struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Year        int       `json:"year"`
	Month       int       `json:"month"`
	CubicMeters float64   `json:"cubic_meters"`
	CreatedAt   time.Time `json:"created_at"`
}

func main() {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// fallback to individual env vars
		host := os.Getenv("PGHOST")
		if host == "" {
			host = "db"
		}
		port := os.Getenv("PGPORT")
		if port == "" {
			port = "5432"
		}
		user := os.Getenv("PGUSER")
		if user == "" {
			user = "postgres"
		}
		password := os.Getenv("PGPASSWORD")
		dbname := os.Getenv("PGDATABASE")
		if dbname == "" {
			dbname = "cagece"
		}
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	app := &App{db: pool}

	if err := app.ensureSchema(ctx); err != nil {
		log.Fatalf("failed to init schema: %v", err)
	}

	r := chi.NewRouter()

	// simple CORS middleware to allow requests from frontend dev server
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("water consumption service"))
	})

	r.Post("/users", app.createUser)
	r.Get("/users", app.listUsers)
	r.Get("/users/{id}", app.getUser)

	r.Post("/users/{id}/consumptions", app.createConsumption)
	r.Get("/users/{id}/consumptions", app.listUserConsumptions)

	r.Get("/consumptions", app.listConsumptions)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	addr := ":" + port
	log.Printf("listening on %s", addr)
	http.ListenAndServe(addr, r)
}

func (a *App) ensureSchema(ctx context.Context) error {
	// execute simple schema creation if not exists
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
	);

	CREATE TABLE IF NOT EXISTS monthly_consumptions (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		year INTEGER NOT NULL,
		month INTEGER NOT NULL,
		cubic_meters NUMERIC NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
		UNIQUE(user_id, year, month)
	);
	`
	_, err := a.db.Exec(ctx, schema)
	return err
}

func respondJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, code int, err string) {
	respondJSON(w, code, map[string]string{"error": err})
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "email required")
		return
	}
	ctx := r.Context()
	var id int
	var created time.Time
	err := a.db.QueryRow(ctx, "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, created_at", req.Name, req.Email).Scan(&id, &created)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	u := User{ID: id, Name: req.Name, Email: req.Email, CreatedAt: created}
	respondJSON(w, http.StatusCreated, u)
}

func (a *App) listUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rows, err := a.db.Query(ctx, "SELECT id, name, email, created_at FROM users ORDER BY id")
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	res := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		res = append(res, u)
	}
	respondJSON(w, http.StatusOK, res)
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ctx := r.Context()
	var u User
	err = a.db.QueryRow(ctx, "SELECT id, name, email, created_at FROM users WHERE id=$1", id).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, u)
}

func (a *App) createConsumption(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	uid, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var req struct {
		Year        int     `json:"year"`
		Month       int     `json:"month"`
		CubicMeters float64 `json:"cubic_meters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Month < 1 || req.Month > 12 {
		respondError(w, http.StatusBadRequest, "month must be 1-12")
		return
	}
	ctx := r.Context()
	var id int
	var created time.Time
	err = a.db.QueryRow(ctx, "INSERT INTO monthly_consumptions (user_id, year, month, cubic_meters) VALUES ($1,$2,$3,$4) RETURNING id, created_at", uid, req.Year, req.Month, req.CubicMeters).Scan(&id, &created)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	c := Consumption{ID: id, UserID: uid, Year: req.Year, Month: req.Month, CubicMeters: req.CubicMeters, CreatedAt: created}
	respondJSON(w, http.StatusCreated, c)
}

func (a *App) listUserConsumptions(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	uid, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	ctx := r.Context()
	rows, err := a.db.Query(ctx, "SELECT id, user_id, year, month, cubic_meters, created_at FROM monthly_consumptions WHERE user_id=$1 ORDER BY year, month", uid)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	res := []Consumption{}
	for rows.Next() {
		var c Consumption
		if err := rows.Scan(&c.ID, &c.UserID, &c.Year, &c.Month, &c.CubicMeters, &c.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		res = append(res, c)
	}
	respondJSON(w, http.StatusOK, res)
}
func (a *App) listConsumptions(w http.ResponseWriter, r *http.Request) {
	// optional query params: user_id, year, month
	q := r.URL.Query()
	userID := q.Get("user_id")
	year := q.Get("year")
	month := q.Get("month")
	ctx := r.Context()

	base := "SELECT id, user_id, year, month, cubic_meters, created_at FROM monthly_consumptions"
	var conds []string
	var args []any
	idx := 1
	if userID != "" {
		conds = append(conds, fmt.Sprintf("user_id=$%d", idx))
		args = append(args, userID)
		idx++
	}
	if year != "" {
		conds = append(conds, fmt.Sprintf("year=$%d", idx))
		args = append(args, year)
		idx++
	}
	if month != "" {
		conds = append(conds, fmt.Sprintf("month=$%d", idx))
		args = append(args, month)
		idx++
	}
	if len(conds) > 0 {
		base = base + " WHERE " + conds[0]
		for i := 1; i < len(conds); i++ {
			base = base + " AND " + conds[i]
		}
	}
	base = base + " ORDER BY year, month"
	rows, err := a.db.Query(ctx, base, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	res := []Consumption{}
	for rows.Next() {
		var c Consumption
		if err := rows.Scan(&c.ID, &c.UserID, &c.Year, &c.Month, &c.CubicMeters, &c.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		res = append(res, c)
	}
	respondJSON(w, http.StatusOK, res)
}
