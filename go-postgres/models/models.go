package models

type Employee struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Dob     string `json:"dob"`
	Country string `json:"country"`
}
