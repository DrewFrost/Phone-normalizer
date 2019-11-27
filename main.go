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

type phoneNumber struct {
	id     int
	number string
}

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

	db, err := connectToExistingDB()
	must(err)
	defer db.Close()
	// numbers := []string{"1234567890", "123 456 7891", "(123) 456 7892", "(123) 456-7893", "123-456-7894", "123-456-7890", "1234567892", "(123)456-7892"}
	// must(createPhoneTable(db))
	// for _, v := range numbers {
	// 	_, err := insertPhone(db, v)
	// 	must(err)
	// }

	phones, err := getAllPhones(db)
	must(err)
	for _, p := range phones {
		fmt.Printf("Working on: %v\n", p)
		phone := normalize(p.number)
		if phone != p.number {
			fmt.Println("Updating or removing: ", phone)
			existing, err := findPhone(db, phone)
			must(err)
			if existing != nil {
				must(deletePhone(db, p.id))
			} else {
				p.number = phone
				must(updatePhone(db, p))
			}
		} else {
			fmt.Println("No changes required")
		}
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
	qr := `
	CREATE TABLE IF NOT EXISTS	phone_numbers (
		id SERIAL,
		value VARCHAR(255)
	)`
	_, err := db.Exec(qr)
	return err
}

func insertPhone(db *sql.DB, phone string) (int, error) {
	var id int
	qr := `INSERT INTO phone_numbers(value) VALUES($1) RETURNING id`
	err := db.QueryRow(qr, phone).Scan(&id)
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

func updatePhone(db *sql.DB, p phoneNumber) error {
	qr := `UPDATE phone_numbers SET value=$2 WHERE id=$1`
	_, err := db.Exec(qr, p.id, p.number)
	return err
}

func deletePhone(db *sql.DB, id int) error {
	qr := `DELETE FROM phone_numbers WHERE id=$1`
	_, err := db.Exec(qr, id)
	return err
}

func findPhone(db *sql.DB, phone string) (*phoneNumber, error) {
	var p phoneNumber
	err := db.QueryRow("SELECT id, value FROM phone_numbers WHERE value=$1", phone).Scan(&p.id, &p.number)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func getAllPhones(db *sql.DB) ([]phoneNumber, error) {
	rows, err := db.Query("SELECT id, value FROM phone_numbers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ret []phoneNumber
	for rows.Next() {
		var p phoneNumber
		if err := rows.Scan(&p.id, &p.number); err != nil {
			return nil, err
		}
		ret = append(ret, p)
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
