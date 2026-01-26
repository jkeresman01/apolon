package main

//go:generate go run github.com/jkeresman01/apolon/apolon-cli generate -i . -o .

// Patient represents a patient record in the database
type Patient struct {
	ID   int    `apolon:"id,pk"`
	Name string `apolon:"name"`
	Age  int    `apolon:"age"`
}
