package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fjl/go-couchdb"
	"github.com/skip2/go-qrcode"
)

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

func generateAndSaveQRCode(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	studentID := r.URL.Query().Get("id")
	if studentID == "" {
		http.Error(w, "Student ID missing", http.StatusBadRequest)
		return
	}

	// Fetch student data from CouchDB
	var student map[string]interface{}
	err := client.DB("student_db").Get(studentID, &student, couchdb.Options{})
	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	// Exclude the _rev field and create a new object for QR code data
	studentDataForQR := map[string]interface{}{
		"id":     student["_id"],
		"name":   student["name"],
		"email":  student["email"],
		"age":    student["age"],
		"gender": student["gender"],
	}

	// Convert the student data (without _rev) to JSON for the QR code
	qrData, err := json.Marshal(studentDataForQR)
	if err != nil {
		http.Error(w, "Failed to marshal student data for QR code", http.StatusInternalServerError)
		return
	}

	// Generate the QR code based on the student details
	qrCodeFile := fmt.Sprintf("%s_qrcode.png", studentID)
	err = qrcode.WriteFile(string(qrData), qrcode.Medium, 256, qrCodeFile)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	// Convert the QR code image to Base64
	file, err := os.ReadFile(qrCodeFile)
	if err != nil {
		http.Error(w, "Failed to read QR code file", http.StatusInternalServerError)
		return
	}
	qrCodeBase64 := base64.StdEncoding.EncodeToString(file)

	// Add the Base64 QR code to the student document
	student["qr_code"] = qrCodeBase64

	// Update the student document in the database
	_, err = client.DB("student_db").Put(studentID, student, student["_rev"].(string))
	if err != nil {
		http.Error(w, "Failed to save QR code in the database", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "QR code generated and saved successfully",
		"qr_code": qrCodeBase64,
	})
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

	http.HandleFunc("/generate-qrcode", func(w http.ResponseWriter, r *http.Request) {
		generateAndSaveQRCode(w, r, client)
	})

	log.Println("Server started at :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

//this code is working perfectly fine
