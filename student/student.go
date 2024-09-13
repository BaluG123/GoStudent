// // package main

// // import (
// // 	"fmt"
// // 	"log"

// // 	"github.com/skip2/go-qrcode"
// // )

// // func main() {
// // 	// Text or URL to encode in the QR code
// // 	data := "https://example.com"

// // 	// Generate the QR code and save it to a PNG file
// // 	err := qrcode.WriteFile(data, qrcode.Medium, 256, "qrcode.png")
// // 	if err != nil {
// // 		log.Fatal(err)
// // 	}

// // 	fmt.Println("QR code generated and saved as qrcode.png")
// // }

// package main

// import (
// 	"encoding/json"
// 	"log"
// 	"net/http"
// 	"sync"

// 	"github.com/fjl/go-couchdb"
// )

// var mutex sync.Mutex

// // CRUD operations for students

// // Create a new student
// func createStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
// 	var student struct {
// 		ID     string `json:"id"`
// 		Name   string `json:"name"`
// 		Email  string `json:"email"`
// 		Age    int    `json:"age"`
// 		Gender string `json:"gender"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
// 		http.Error(w, "invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	var existingDoc map[string]interface{}
// 	err := client.DB("student_db").Get(student.ID, &existingDoc, couchdb.Options{})
// 	if err == nil {
// 		http.Error(w, "Student ID already exists", http.StatusConflict)
// 		return
// 	}

// 	doc := map[string]interface{}{
// 		"_id":    student.ID,
// 		"name":   student.Name,
// 		"email":  student.Email,
// 		"age":    student.Age,
// 		"gender": student.Gender,
// 	}

// 	_, err = client.DB("student_db").Put(student.ID, doc, "")
// 	if err != nil {
// 		http.Error(w, "failed to create student", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(map[string]string{"message": "Student created successfully"})
// }

// // Retrieve a student by ID
// func getStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
// 	studentID := r.URL.Query().Get("id")
// 	if studentID == "" {
// 		http.Error(w, "Student ID missing", http.StatusBadRequest)
// 		return
// 	}

// 	var student map[string]interface{}
// 	err := client.DB("student_db").Get(studentID, &student, couchdb.Options{})
// 	if err != nil {
// 		http.Error(w, "Student not found", http.StatusNotFound)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(student)
// }

// // Update student data
// func updateStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
// 	var student struct {
// 		ID     string `json:"id"`
// 		Name   string `json:"name"`
// 		Email  string `json:"email"`
// 		Age    int    `json:"age"`
// 		Gender string `json:"gender"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
// 		http.Error(w, "invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	var existingDoc map[string]interface{}
// 	err := client.DB("student_db").Get(student.ID, &existingDoc, couchdb.Options{})
// 	if err != nil {
// 		http.Error(w, "Student not found", http.StatusNotFound)
// 		return
// 	}

// 	doc := map[string]interface{}{
// 		"_id":    student.ID,
// 		"_rev":   existingDoc["_rev"],
// 		"name":   student.Name,
// 		"email":  student.Email,
// 		"age":    student.Age,
// 		"gender": student.Gender,
// 	}

// 	_, err = client.DB("student_db").Put(student.ID, doc, existingDoc["_rev"].(string))
// 	if err != nil {
// 		http.Error(w, "failed to update student", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(map[string]string{"message": "Student updated successfully"})
// }

// // Delete student by ID
// func deleteStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
// 	studentID := r.URL.Query().Get("id")
// 	if studentID == "" {
// 		http.Error(w, "Student ID missing", http.StatusBadRequest)
// 		return
// 	}

// 	var existingDoc map[string]interface{}
// 	err := client.DB("student_db").Get(studentID, &existingDoc, couchdb.Options{})
// 	if err != nil {
// 		http.Error(w, "Student not found", http.StatusNotFound)
// 		return
// 	}

// 	_, err = client.DB("student_db").Delete(studentID, existingDoc["_rev"].(string))
// 	if err != nil {
// 		http.Error(w, "failed to delete student", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(map[string]string{"message": "Student deleted successfully"})
// }

// // CouchDB connection function
// func initCouchDB() (*couchdb.Client, error) {
// 	client, err := couchdb.NewClient("http://admin:admin@localhost:5984/", nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return client, nil
// }

// func main() {
// 	client, err := initCouchDB()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	http.HandleFunc("/create-student", func(w http.ResponseWriter, r *http.Request) {
// 		createStudent(w, r, client)
// 	})

// 	http.HandleFunc("/get-student", func(w http.ResponseWriter, r *http.Request) {
// 		getStudent(w, r, client)
// 	})

// 	http.HandleFunc("/update-student", func(w http.ResponseWriter, r *http.Request) {
// 		updateStudent(w, r, client)
// 	})

// 	http.HandleFunc("/delete-student", func(w http.ResponseWriter, r *http.Request) {
// 		deleteStudent(w, r, client)
// 	})

// 	log.Println("Server started at :8081")
// 	log.Fatal(http.ListenAndServe(":8081", nil))
// }

// //this code is working perfectly fine

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fjl/go-couchdb"
	"github.com/skip2/go-qrcode"
)

type CustomTime struct {
	time.Time
}

const ctLayout = "2006-01-02"

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(`"`+ctLayout+`"`, s)
	return
}

// Student struct
type Student struct {
	ID                        string             `json:"id"`
	FullName                  string             `json:"full_name"`
	DateOfBirth               CustomTime         `json:"date_of_birth"`
	Gender                    string             `json:"gender"`
	Address                   string             `json:"address"`
	ContactNumber             string             `json:"contact_number"`
	EmailAddress              string             `json:"email_address"`
	EmergencyContact          string             `json:"emergency_contact"`
	Class                     string             `json:"class"`
	Section                   string             `json:"section"`
	RollNumber                string             `json:"roll_number"`
	SubjectsEnrolled          []string           `json:"subjects_enrolled"`
	AttendanceRecords         []AttendanceRecord `json:"attendance_records"`
	ExamScores                []ExamScore        `json:"exam_scores"`
	ExtracurricularActivities []string           `json:"extracurricular_activities"`
	BehavioralRecords         []BehavioralRecord `json:"behavioral_records"`
	HealthRecords             []HealthRecord     `json:"health_records"`
	AdmissionDate             CustomTime         `json:"admission_date"`
	PreviousSchool            string             `json:"previous_school"`
	FeePaymentRecords         []FeePaymentRecord `json:"fee_payment_records"`
	Scholarships              []Scholarship      `json:"scholarships"`
}

// AttendanceRecord struct
type AttendanceRecord struct {
	Date   CustomTime `json:"date"`
	Status string     `json:"status"`
}

// ExamScore struct
type ExamScore struct {
	Subject string  `json:"subject"`
	Score   float64 `json:"score"`
	Grade   string  `json:"grade"`
}

// BehavioralRecord struct
type BehavioralRecord struct {
	Date        CustomTime `json:"date"`
	Incident    string     `json:"incident"`
	ActionTaken string     `json:"action_taken"`
}

// HealthRecord struct
type HealthRecord struct {
	Condition string `json:"condition"`
	Notes     string `json:"notes"`
}

// FeePaymentRecord struct
type FeePaymentRecord struct {
	Date   CustomTime `json:"date"`
	Amount float64    `json:"amount"`
	Status string     `json:"status"`
}

// Scholarship struct
type Scholarship struct {
	Name        string     `json:"name"`
	Amount      float64    `json:"amount"`
	DateAwarded CustomTime `json:"date_awarded"`
}

// CRUD operations for students

func createStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var student Student

	// Debug: Log the incoming request body
	log.Println("Received request to create student")

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// Debug: Log the raw request body
	log.Printf("Raw request body: %s", string(body))

	// Decode the request body into the student struct
	if err := json.Unmarshal(body, &student); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// Debug: Log the decoded student struct
	log.Printf("Decoded student: %+v", student)

	var existingDoc map[string]interface{}
	err = client.DB("student_db").Get(student.ID, &existingDoc, couchdb.Options{})
	if err == nil {
		http.Error(w, "Student ID already exists", http.StatusConflict)
		return
	}

	doc := map[string]interface{}{
		"_id":                        student.ID,
		"full_name":                  student.FullName,
		"date_of_birth":              student.DateOfBirth.Format(ctLayout), // Format date for CouchDB
		"gender":                     student.Gender,
		"address":                    student.Address,
		"contact_number":             student.ContactNumber,
		"email_address":              student.EmailAddress,
		"emergency_contact":          student.EmergencyContact,
		"class":                      student.Class,
		"section":                    student.Section,
		"roll_number":                student.RollNumber,
		"subjects_enrolled":          student.SubjectsEnrolled,
		"attendance_records":         student.AttendanceRecords,
		"exam_scores":                student.ExamScores,
		"extracurricular_activities": student.ExtracurricularActivities,
		"behavioral_records":         student.BehavioralRecords,
		"health_records":             student.HealthRecords,
		"admission_date":             student.AdmissionDate.Format(ctLayout), // Format date for CouchDB
		"previous_school":            student.PreviousSchool,
		"fee_payment_records":        student.FeePaymentRecords,
		"scholarships":               student.Scholarships,
	}

	_, err = client.DB("student_db").Put(student.ID, doc, "")
	if err != nil {
		http.Error(w, "failed to create student", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Student created successfully"})
}

// Create a new student
// func createStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
// 	var student Student

// 	log.Println("Received request to create student", r.Body)

// 	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
// 		http.Error(w, "invalid request payload", http.StatusBadRequest)
// 		return
// 	}

// 	log.Printf("Decoded student: %+v", student)

// 	var existingDoc map[string]interface{}
// 	err := client.DB("student_db").Get(student.ID, &existingDoc, couchdb.Options{})
// 	if err == nil {
// 		http.Error(w, "Student ID already exists", http.StatusConflict)
// 		return
// 	}

// 	doc := map[string]interface{}{
// 		"_id":                        student.ID,
// 		"full_name":                  student.FullName,
// 		"date_of_birth":              student.DateOfBirth.Format("2006-01-02"), // Format date for CouchDB
// 		"gender":                     student.Gender,
// 		"address":                    student.Address,
// 		"contact_number":             student.ContactNumber,
// 		"email_address":              student.EmailAddress,
// 		"emergency_contact":          student.EmergencyContact,
// 		"class":                      student.Class,
// 		"section":                    student.Section,
// 		"roll_number":                student.RollNumber,
// 		"subjects_enrolled":          student.SubjectsEnrolled,
// 		"attendance_records":         student.AttendanceRecords,
// 		"exam_scores":                student.ExamScores,
// 		"extracurricular_activities": student.ExtracurricularActivities,
// 		"behavioral_records":         student.BehavioralRecords,
// 		"health_records":             student.HealthRecords,
// 		"admission_date":             student.AdmissionDate.Format("2006-01-02"), // Format date for CouchDB
// 		"previous_school":            student.PreviousSchool,
// 		"fee_payment_records":        student.FeePaymentRecords,
// 		"scholarships":               student.Scholarships,
// 	}

// 	_, err = client.DB("student_db").Put(student.ID, doc, "")
// 	if err != nil {
// 		http.Error(w, "failed to create student", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(map[string]string{"message": "Student created successfully"})
// }

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

	// Create a map to store the data for the QR code
	studentDataForQR := map[string]interface{}{
		"id":             student["_id"],
		"full_name":      student["full_name"],
		"contact_number": student["contact_number"],
		"email_address":  student["email_address"],
		"class":          student["class"],
	}

	// Convert the QR code data to JSON
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

	var student Student
	err := client.DB("student_db").Get(studentID, &student, couchdb.Options{})
	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(student)
}

// Update student data
// Update student data (continued)
func updateStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var student Student

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
		"_id":                        student.ID,
		"_rev":                       existingDoc["_rev"],
		"full_name":                  student.FullName,
		"date_of_birth":              student.DateOfBirth.Format("2006-01-02"), // Format date for CouchDB
		"gender":                     student.Gender,
		"address":                    student.Address,
		"contact_number":             student.ContactNumber,
		"email_address":              student.EmailAddress,
		"emergency_contact":          student.EmergencyContact,
		"class":                      student.Class,
		"section":                    student.Section,
		"roll_number":                student.RollNumber,
		"subjects_enrolled":          student.SubjectsEnrolled,
		"attendance_records":         student.AttendanceRecords,
		"exam_scores":                student.ExamScores,
		"extracurricular_activities": student.ExtracurricularActivities,
		"behavioral_records":         student.BehavioralRecords,
		"health_records":             student.HealthRecords,
		"admission_date":             student.AdmissionDate.Format("2006-01-02"), // Format date for CouchDB
		"previous_school":            student.PreviousSchool,
		"fee_payment_records":        student.FeePaymentRecords,
		"scholarships":               student.Scholarships,
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
