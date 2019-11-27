package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	host   = "localhost"
	port   = 5432
	dbname = "number_normalizer"
)

var user, password string

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	user = os.Getenv("POSTGRES_USER")
	password = os.Getenv("POSTGRES_PASS")
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", host, port, user, password)
	db, err := sql.Open("postgres", psqlInfo)
	must(err)
	err = resetDB(db, dbname)
	must(err)
	db.Close()
	psqlInfo = fmt.Sprintf("%s dbname=%s", psqlInfo, dbname)
	db, err = sql.Open("postgres", psqlInfo)
	must(err)
	defer db.Close()

	must(createPhoneTable(db))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func createPhoneTable(db *sql.DB) error {
	statement := `
	CREATE TABLE IF NOT EXISTS	phone_numbers (
		id SERIAL,
		value VARCHAR(255)
	)`
	_, err := db.Exec(statement)
	return err
}

func resetDB(db *sql.DB, name string) error {
	qr := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbname)
	_, err := db.Exec(qr)
	if err != nil {
		return err
	}
	return createDB(db, name)
}

func createDB(db *sql.DB, name string) error {
	qr := fmt.Sprintf("CREATE DATABASE %s", dbname)
	_, err := db.Exec(qr)
	if err != nil {
		return err
	}
	return nil
}

//Regex
func normalize(phone string) string {
	re := regexp.MustCompile("[^0-9]")
	return re.ReplaceAllString(phone, "")
}

// Normaliza Without Regex
// func normalize(phone string) string {
// 	var buf bytes.Buffer
// 	for _, ch := range phone {
// 		if ch >= '1' && ch <= '9' {
// 			buf.WriteRune(ch)
// 		}
// 	}
// 	return buf.String()
// }
