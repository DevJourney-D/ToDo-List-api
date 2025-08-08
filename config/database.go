package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var DB *sql.DB

type Config struct {
	DatabaseURL string
	SupabaseURL string
	SupabaseKey string
	JWTSecret   string
	Port        string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", ""),
		SupabaseURL: getEnv("SUPABASE_URL", ""),
		SupabaseKey: getEnv("SUPABASE_ANON_KEY", ""),
		JWTSecret:   getEnv("JWT_SECRET", "default-secret"),
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func InitDatabase(databaseURL string) {
	var err error
	DB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool for better performance
	DB.SetMaxOpenConns(25)                 // Maximum number of open connections
	DB.SetMaxIdleConns(10)                 // Maximum number of idle connections
	DB.SetConnMaxLifetime(5 * time.Minute) // Maximum time a connection can be reused

	if err = DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Successfully connected to database with optimized pool settings")
}

func CloseDatabase() {
	if DB != nil {
		DB.Close()
	}
}
