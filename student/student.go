// package main

// import (
// 	"fmt"
// 	"log"

// 	"github.com/skip2/go-qrcode"
// )

// func main() {
// 	// Text or URL to encode in the QR code
// 	data := "https://example.com"

// 	// Generate the QR code and save it to a PNG file
// 	err := qrcode.WriteFile(data, qrcode.Medium, 256, "qrcode.png")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Println("QR code generated and saved as qrcode.png")
// }

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/fjl/go-couchdb"
)

var mutex sync.Mutex

// CRUD operations for students

// Create a new student
func createStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var student struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Age    int    `json:"age"`
		Gender string `json:"gender"`
	}

	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	var existingDoc map[string]interface{}
	err := client.DB("student_db").Get(student.ID, &existingDoc, couchdb.Options{})
	if err == nil {
		http.Error(w, "Student ID already exists", http.StatusConflict)
		return
	}

	doc := map[string]interface{}{
		"_id":    student.ID,
		"name":   student.Name,
		"email":  student.Email,
		"age":    student.Age,
		"gender": student.Gender,
	}

	_, err = client.DB("student_db").Put(student.ID, doc, "")
	if err != nil {
		http.Error(w, "failed to create student", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Student created successfully"})
}

// Retrieve a student by ID
func getStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	studentID := r.URL.Query().Get("id")
	if studentID == "" {
		http.Error(w, "Student ID missing", http.StatusBadRequest)
		return
	}

	var student map[string]interface{}
	err := client.DB("student_db").Get(studentID, &student, couchdb.Options{})
	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(student)
}

// Update student data
func updateStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var student struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Age    int    `json:"age"`
		Gender string `json:"gender"`
	}

	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	var existingDoc map[string]interface{}
	err := client.DB("student_db").Get(student.ID, &existingDoc, couchdb.Options{})
	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	doc := map[string]interface{}{
		"_id":    student.ID,
		"_rev":   existingDoc["_rev"],
		"name":   student.Name,
		"email":  student.Email,
		"age":    student.Age,
		"gender": student.Gender,
	}

	_, err = client.DB("student_db").Put(student.ID, doc, existingDoc["_rev"].(string))
	if err != nil {
		http.Error(w, "failed to update student", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Student updated successfully"})
}

// Delete student by ID
func deleteStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	studentID := r.URL.Query().Get("id")
	if studentID == "" {
		http.Error(w, "Student ID missing", http.StatusBadRequest)
		return
	}

	var existingDoc map[string]interface{}
	err := client.DB("student_db").Get(studentID, &existingDoc, couchdb.Options{})
	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	_, err = client.DB("student_db").Delete(studentID, existingDoc["_rev"].(string))
	if err != nil {
		http.Error(w, "failed to delete student", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Student deleted successfully"})
}

// CouchDB connection function
func initCouchDB() (*couchdb.Client, error) {
	client, err := couchdb.NewClient("http://admin:admin@localhost:5984/", nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func main() {
	client, err := initCouchDB()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/create-student", func(w http.ResponseWriter, r *http.Request) {
		createStudent(w, r, client)
	})

	http.HandleFunc("/get-student", func(w http.ResponseWriter, r *http.Request) {
		getStudent(w, r, client)
	})

	http.HandleFunc("/update-student", func(w http.ResponseWriter, r *http.Request) {
		updateStudent(w, r, client)
	})

	http.HandleFunc("/delete-student", func(w http.ResponseWriter, r *http.Request) {
		deleteStudent(w, r, client)
	})

	log.Println("Server started at :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

//this code is working perfectly fine
