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
	// initDB()
	// numbers := []string{"1234567890", "123 456 7891", "(123) 456 7892", "(123) 456-7893", "123-456-7894", "123-456-7890", "1234567892", "(123)456-7892"}
	db, err := connectToExistingDB()
	must(err)
	defer db.Close()
	// must(createPhoneTable(db))
	// for i, v := range numbers {
	// 	id, err := insertPhone(db, v)
	// 	must(err)
	// 	ids[i] = id
	// }
	phones, err := getAllPhones(db)
	must(err)
	for _, p := range phones {
		fmt.Printf("%v\n", p)
	}

}

func initDB() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", host, port, user, password)
	db, err := sql.Open("postgres", psqlInfo)
	must(err)
	err = resetDB(db, dbname)
	must(err)
	db.Close()
}

func connectToExistingDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	must(err)
	return db, nil
}

func createDB(db *sql.DB, name string) error {
	qr := fmt.Sprintf("CREATE DATABASE %s", dbname)
	_, err := db.Exec(qr)
	if err != nil {
		return err
	}
	return nil
}

func resetDB(db *sql.DB, name string) error {
	qr := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbname)
	_, err := db.Exec(qr)
	if err != nil {
		return err
	}
	return createDB(db, name)
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

func insertPhone(db *sql.DB, phone string) (int, error) {
	var id int
	statement := `INSERT INTO phone_numbers(value) VALUES($1) RETURNING id`
	err := db.QueryRow(statement, phone).Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, nil
}

func getPhone(db *sql.DB, id int) (string, error) {
	var phone string
	err := db.QueryRow("SELECT value FROM phone_numbers WHERE id=$1", id).Scan(&phone)
	if err != nil {
		return "", err
	}
	return phone, nil
}

func getAllPhones(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT value FROM phone_numbers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []string
	for rows.Next() {
		var phone string
		if err := rows.Scan(&phone); err != nil {
			return nil, err
		}
		ret = append(ret, phone)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
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
