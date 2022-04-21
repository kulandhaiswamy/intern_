package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-postgres/models"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type response struct {
	Id      int    `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

func createConnection() *sql.DB {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))

	if err != nil {
		panic(err)
	}

	err = db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected")

	return db
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var user models.Employee

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Fatalf("Unable to decode the request body %v", err)
	}

	insertID := insertUser(user)

	res := response{
		Id:      insertID,
		Message: "User created successfully",
	}

	json.NewEncoder(w).Encode(res)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert string into %v", err)

	}

	user, err := getUser(int(id))

	if err != nil {
		log.Fatalf("Unable to get user %v", err)
	}

	json.NewEncoder(w).Encode(user)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int %v", err)

	}

	var user models.Employee

	err = json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Fatalf("Unable to decode the request body %v", err)
	}

	updatedRows := updateUser(int(id), user)

	msg := fmt.Sprintf("User updated succesfully .total rows affected %v", updatedRows)

	res := response{
		Id:      int(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)

}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Context-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int %v", err)
	}

	deletedRows := deleteUser(int(id))

	msg := fmt.Sprintf("User deleted successfully.Total rows affected %v", deletedRows)

	res := response{
		Id:      int(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

func updateUser(id int, user models.Employee) int {
	db := createConnection()

	defer db.Close()

	sqlStatement := `UPDATE employee SET name=$2,dob=$3,country=$4 WHERE id=$1`

	res, err := db.Exec(sqlStatement, id, user.Name, user.Dob, user.Country)

	if err != nil {
		log.Fatalf("Unable to execute query %v", err)
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows %v", err)

	}

	fmt.Printf("Total rows affected %v", rowsAffected)

	return int(rowsAffected)
}

func insertUser(user models.Employee) int {

	db := createConnection()

	defer db.Close()

	sqlStatement := `INSERT INTO employee(id,name,dob,country) VALUES($1,$2,$3,$4) RETURNING id`

	var id int

	err := db.QueryRow(sqlStatement, user.Id, user.Name, user.Dob, user.Country).Scan(&id)

	if err != nil {
		log.Fatalf("Unable to execute query %v", err)
	}

	fmt.Printf("\nInserted a single record %v", id)

	return id
}

func getUser(id int) (models.Employee, error) {
	db := createConnection()

	defer db.Close()

	var user models.Employee

	sqlStatement := `SELECT * FROM employee WHERE id=$1`

	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(&user.Id, &user.Name, &user.Dob, &user.Country)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned")
		return user, nil
	case nil:
		return user, nil
	default:
		log.Fatalf("Unable to scan the row %v", err)
	}
	return user, err

}

func deleteUser(id int) int {
	db := createConnection()

	defer db.Close()

	sqlStatement := `DELETE FROM employee WHERE id=$1`

	res, err := db.Exec(sqlStatement, id)

	if err != nil {
		log.Fatalf("Unable to execute query %v", err)
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows %v", err)
	}

	fmt.Printf("Total rows affected %v", rowsAffected)

	return int(rowsAffected)

}
