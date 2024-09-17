package teacher

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fjl/go-couchdb"
	"github.com/skip2/go-qrcode"
)

type CustomTime struct {
	time.Time
}

const ctLayout = "2006-01-02"

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		ct.Time = time.Time{}
		return
	}
	ct.Time, err = time.Parse(ctLayout, s)
	return
}

type Teacher struct {
	ID             string          `json:"id"`
	FullName       string          `json:"full_name"`
	DateOfBirth    CustomTime      `json:"date_of_birth"`
	Gender         string          `json:"gender"`
	Address        string          `json:"address"`
	ContactNumber  string          `json:"contact_number"`
	EmailAddress   string          `json:"email_address"`
	Department     string          `json:"department"`
	SubjectsTaught []string        `json:"subjects_taught"`
	Qualification  []Qualification `json:"qualification"`
	Experience     int             `json:"experience"`
	JoiningDate    CustomTime      `json:"joining_date"`
	PreviousSchool string          `json:"previous_school"`
	Salary         float64         `json:"salary"`
	LeaveRecords   []LeaveRecord   `json:"leave_records"`
}

type Qualification struct {
	Degree    string `json:"degree"`
	Major     string `json:"major"`
	Year      int    `json:"year"`
	Institute string `json:"institute"`
}

type LeaveRecord struct {
	StartDate CustomTime `json:"start_date"`
	EndDate   CustomTime `json:"end_date"`
	Reason    string     `json:"reason"`
	Status    string     `json:"status"`
}

func CreateTeacher(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var teacher Teacher

	log.Println("Received request to create teacher")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Raw request body: %s", string(body))

	if err := json.Unmarshal(body, &teacher); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Decoded Teacher: %+v", teacher)

	var existingDoc map[string]interface{}
	err = client.DB("teacher_db").Get(teacher.ID, &existingDoc, couchdb.Options{})
	if err == nil {
		http.Error(w, "Teacher ID already exists", http.StatusConflict)
		return
	}

	doc := map[string]interface{}{
		"_id":             teacher.ID,
		"full_name":       teacher.FullName,
		"date_of_birth":   teacher.DateOfBirth.Format(ctLayout),
		"gender":          teacher.Gender,
		"address":         teacher.Address,
		"contact_number":  teacher.ContactNumber,
		"email_address":   teacher.EmailAddress,
		"department":      teacher.Department,
		"subjects_taught": teacher.SubjectsTaught,
		"qualification":   teacher.Qualification,
		"experience":      teacher.Experience,
		"joining_date":    teacher.JoiningDate.Format(ctLayout),
		"previous_school": teacher.PreviousSchool,
		"salary":          teacher.Salary,
		"leave_records":   teacher.LeaveRecords,
	}

	_, err = client.DB("teacher_db").Put(teacher.ID, doc, "")
	if err != nil {
		http.Error(w, "failed to create teacher", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Teacher created successfully"})
}

func GenerateAndSaveQRCode(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	teacherID := r.URL.Query().Get("id")
	if teacherID == "" {
		http.Error(w, "Teacher ID missing", http.StatusBadRequest)
		return
	}

	// Fetch student data from CouchDB
	var teacher map[string]interface{}
	err := client.DB("teacher_db").Get(teacherID, &teacher, couchdb.Options{})
	if err != nil {
		http.Error(w, "tecaher not found", http.StatusNotFound)
		return
	}

	// Create a map to store the data for the QR code
	teacherDataForQR := map[string]interface{}{
		"id":              teacher["_id"],
		"full_name":       teacher["full_name"],
		"department":      teacher["contact_number"],
		"email_address":   teacher["email_address"],
		"subjects_taught": teacher["subjects_taught"],
	}

	// Convert the QR code data to JSON
	qrData, err := json.Marshal(teacherDataForQR)
	if err != nil {
		http.Error(w, "Failed to marshal student data for QR code", http.StatusInternalServerError)
		return
	}

	// Generate the QR code in memory
	qrCode, err := qrcode.Encode(string(qrData), qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	// Convert the QR code image to Base64
	qrCodeBase64 := base64.StdEncoding.EncodeToString(qrCode)

	// Add the Base64 QR code to the student document
	teacher["qr_code"] = qrCodeBase64

	// Update the student document in the database
	_, err = client.DB("teacher_db").Put(teacherID, teacher, teacher["_rev"].(string))
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
func GetTeacher(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	teacherID := r.URL.Query().Get("id")
	if teacherID == "" {
		http.Error(w, "teacher ID missing", http.StatusBadRequest)
		return
	}

	// var student Student
	var teacher map[string]interface{}
	err := client.DB("teacher_db").Get(teacherID, &teacher, couchdb.Options{})
	fmt.Println("api got hit")
	if err != nil {
		http.Error(w, "teacher not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(teacher)
}

func GetAllTeachers(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	// Define a struct for the result of AllDocs
	type AllDocsResult struct {
		Rows []struct {
			ID  string          `json:"id"`
			Doc json.RawMessage `json:"doc"`
		} `json:"rows"`
	}

	var result AllDocsResult

	// Fetch all documents (students) from the "student_db"
	err := client.DB("teacher_db").AllDocs(&result, couchdb.Options{
		"include_docs": true, // Include full documents, not just IDs
	})
	if err != nil {
		http.Error(w, "Failed to fetch students", http.StatusInternalServerError)
		return
	}

	// Prepare a slice to store the student details
	var teachers []map[string]interface{}

	// Iterate over the rows and append the student details to the slice
	for _, row := range result.Rows {
		var teacher map[string]interface{}
		if err := json.Unmarshal(row.Doc, &teacher); err == nil {
			// Exclude the _rev field if necessary
			delete(teacher, "_rev")
			teachers = append(teachers, teacher)
		}
	}

	// Respond with the list of students
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(teachers)
}

func UpdateTeacher(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var teacher Teacher

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// Log the raw request body
	log.Printf("Raw request body: %s", string(body))

	// Decode the request body into the teacher struct
	if err := json.Unmarshal(body, &teacher); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Decoded Teacher: %+v", teacher)

	var existingDoc map[string]interface{}
	err = client.DB("teacher_db").Get(teacher.ID, &existingDoc, couchdb.Options{})
	if err != nil {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	}

	doc := map[string]interface{}{
		"_id":             teacher.ID,
		"_rev":            existingDoc["_rev"],
		"full_name":       teacher.FullName,
		"date_of_birth":   teacher.DateOfBirth.Format(ctLayout),
		"gender":          teacher.Gender,
		"address":         teacher.Address,
		"contact_number":  teacher.ContactNumber,
		"email_address":   teacher.EmailAddress,
		"department":      teacher.Department,
		"subjects_taught": teacher.SubjectsTaught,
		"qualification":   teacher.Qualification,
		"experience":      teacher.Experience,
		"joining_date":    teacher.JoiningDate.Format(ctLayout),
		"previous_school": teacher.PreviousSchool,
		"salary":          teacher.Salary,
		"leave_records":   teacher.LeaveRecords,
	}

	_, err = client.DB("teacher_db").Put(teacher.ID, doc, existingDoc["_rev"].(string))
	if err != nil {
		log.Printf("Error updating teacher: %v", err)
		http.Error(w, "failed to update teacher", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Teacher updated successfully"})
}

// Delete student by ID
func DeleteTeacher(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	teacherID := r.URL.Query().Get("id")
	if teacherID == "" {
		http.Error(w, "Tecaher ID missing", http.StatusBadRequest)
		return
	}

	var existingDoc map[string]interface{}
	err := client.DB("teacher_db").Get(teacherID, &existingDoc, couchdb.Options{})
	if err != nil {
		http.Error(w, "teacher not found", http.StatusNotFound)
		return
	}

	_, err = client.DB("teacher_db").Delete(teacherID, existingDoc["_rev"].(string))
	if err != nil {
		http.Error(w, "failed to delete teacher", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Teacher deleted successfully"})
}
