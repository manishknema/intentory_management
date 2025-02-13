package database

import (
	"database/sql"
	"log"

	_ "github.com/glebarez/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "inventory.db")
	if err != nil {
		log.Fatal("❌ Database connection error:", err)
	}

	log.Println("✅ Database connected!")

	// Create Users Table
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL
        );
    `)
	if err != nil {
		log.Fatal("❌ Users table creation failed:", err)
	}
	log.Println("✅ Users table ready")

	// Create Inventory Table
	_, err = DB.Exec(`
        CREATE TABLE IF NOT EXISTS inventory (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            description TEXT,
            price FLOAT NOT NULL
        );
    `)
	if err != nil {
		log.Fatal("❌ Inventory table creation failed:", err)
	}
	log.Println("✅ Inventory table ready")
}
