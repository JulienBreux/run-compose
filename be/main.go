package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

var (
	ctx = context.Background()
	rdb *redis.Client
	db  *sql.DB
)

type Meal struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Calories  int       `json:"calories"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	// Database connection variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	// Redis connection variables
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	// Connect to Redis
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	// Connect to Postgres
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	var err error
	// Retry logic for DB connection
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", psqlInfo)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			log.Println("Connected to database")
			break
		}
		log.Printf("Failed to connect to database (attempt %d/10): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}

	// Initialize schema
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS meals (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		calories INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/meals", handleMeals)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleMeals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		getMeals(w, r)
	} else if r.Method == "POST" {
		createMeal(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getMeals(w http.ResponseWriter, r *http.Request) {
	// Try to get from cache
	val, err := rdb.Get(ctx, "meals").Result()
	if err == nil {
		log.Println("Cache hit")
		w.Write([]byte(val))
		return
	}

	log.Println("Cache miss")
	rows, err := db.Query("SELECT id, name, calories, created_at FROM meals ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var meals []Meal
	for rows.Next() {
		var m Meal
		if err := rows.Scan(&m.ID, &m.Name, &m.Calories, &m.CreatedAt); err != nil {
			log.Println(err)
			continue
		}
		meals = append(meals, m)
	}

	jsonBytes, err := json.Marshal(meals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Cache result for 10 seconds
	rdb.Set(ctx, "meals", jsonBytes, 10*time.Second)

	w.Write(jsonBytes)
}

func createMeal(w http.ResponseWriter, r *http.Request) {
	var m Meal
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := db.QueryRow("INSERT INTO meals (name, calories) VALUES ($1, $2) RETURNING id, created_at",
		m.Name, m.Calories).Scan(&m.ID, &m.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Invalidate cache
	rdb.Del(ctx, "meals")

	json.NewEncoder(w).Encode(m)
}
