package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/fjl/go-couchdb"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("my_secret_key")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type Student struct {
	ID                        string             `json:"id"`
	FullName                  string             `json:"full_name"`
	DateOfBirth               time.Time          `json:"date_of_birth"`
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
	AdmissionDate             time.Time          `json:"admission_date"`
	PreviousSchool            string             `json:"previous_school"`
	FeePaymentRecords         []FeePaymentRecord `json:"fee_payment_records"`
	Scholarships              []Scholarship      `json:"scholarships"`
}

type AttendanceRecord struct {
	Date   time.Time `json:"date"`
	Status string    `json:"status"`
}

type ExamScore struct {
	Subject string  `json:"subject"`
	Score   float64 `json:"score"`
	Grade   string  `json:"grade"`
}

type BehavioralRecord struct {
	Date        time.Time `json:"date"`
	Incident    string    `json:"incident"`
	ActionTaken string    `json:"action_taken"`
}

type HealthRecord struct {
	Condition string `json:"condition"`
	Notes     string `json:"notes"`
}

type FeePaymentRecord struct {
	Date   time.Time `json:"date"`
	Amount float64   `json:"amount"`
	Status string    `json:"status"`
}

type Scholarship struct {
	Name        string    `json:"name"`
	Amount      float64   `json:"amount"`
	DateAwarded time.Time `json:"date_awarded"`
}

func login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	storedPassword, ok := users[creds.Username]
	if !ok || bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(creds.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Login successful"})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		tokenStr := cookie.Value
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func createStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var student Student
	err := json.NewDecoder(r.Body).Decode(&student)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err = client.DB("student_db").Put(student.ID, student, "")
	if err != nil {
		http.Error(w, "Failed to create student", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Student created successfully"})
}

func getStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	studentID := r.URL.Query().Get("id")
	if studentID == "" {
		http.Error(w, "Student ID missing", http.StatusBadRequest)
		return
	}

	var student Student
	err := client.DB("student_db").Get(studentID, &student, nil)
	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(student)
}

func updateStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var student Student
	err := json.NewDecoder(r.Body).Decode(&student)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	_, err = client.DB("student_db").Put(student.ID, student, "")
	if err != nil {
		http.Error(w, "Failed to update student", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Student updated successfully"})
}

func deleteStudent(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	studentID := r.URL.Query().Get("id")
	if studentID == "" {
		http.Error(w, "Student ID missing", http.StatusBadRequest)
		return
	}

	var student Student
	err := client.DB("student_db").Get(studentID, &student, nil)
	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	_, err = client.DB("student_db").Delete(studentID, student)
	if err != nil {
		http.Error(w, "Failed to delete student", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Student deleted successfully"})
}

func main() {
	client, err := couchdb.NewClient("http://admin:admin@localhost:5984/", nil)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/login", login)
	http.Handle("/create-student", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		createStudent(w, r, client)
	})))
	http.Handle("/get-student", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		getStudent(w, r, client)
	})))
	http.Handle("/update-student", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		updateStudent(w, r, client)
	})))
	http.Handle("/delete-student", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deleteStudent(w, r, client)
	})))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
