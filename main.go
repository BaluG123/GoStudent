package main

import (
	"data-access/student"
	"log"
	"net/http"

	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"

	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/fjl/go-couchdb"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type OTP struct {
	Code   string
	Expiry time.Time
}

var otpStore = make(map[string]OTP)
var mutex sync.Mutex
var jwtSecret = []byte("PUbCXwBRB1M0U9gQwHoi809E8fEYux2U")

type contextKey string

const emailKey contextKey = "email"

func generateOTP() string {
	const otpLength = 6
	const otpChars = "0123456789"
	result := make([]byte, otpLength)
	for i := 0; i < otpLength; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(otpChars))))
		result[i] = otpChars[num.Int64()]
	}
	return string(result)
}

// send otp
func sendOTP(email string, otp string) error {
	from := "oyprasad1432@gmail.com"
	password := "mrrlvdhaxwxkmohu"
	to := email
	subject := "subject: your otp \n"
	body := "your otp code is:" + otp
	msg := subject + "\n" + body

	// smtp authentication object
	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, []byte(msg))
	if err != nil {
		return err
	}
	return nil
}

// request otp
func requestOTP(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	otp := generateOTP()
	expiry := time.Now().Add(100 * time.Minute)

	mutex.Lock()
	otpStore[request.Email] = OTP{Code: otp, Expiry: expiry}
	mutex.Unlock()

	if err := sendOTP(request.Email, otp); err != nil {
		http.Error(w, "failed to send OTP", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("otp send successful"))
}

// Faculty Registration
func registerFaculty(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var request struct {
		ID          string `json:"id"`
		TYPE        string `json:"type"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		Designation string `json:"designation"`
		Password    string `json:"password"`
		OTP         string `json:"otp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	otp, exists := otpStore[request.Email]
	mutex.Unlock()

	if !exists || otp.Code != request.OTP || time.Now().After(otp.Expiry) {
		http.Error(w, "invalid or expired otp", http.StatusUnauthorized)
		return
	}

	mutex.Lock()
	delete(otpStore, request.Email)
	mutex.Unlock()

	var existingDoc map[string]interface{}
	err := client.DB("education_management").Get(request.ID, &existingDoc, couchdb.Options{})
	if err == nil {
		http.Error(w, "ID already exists", http.StatusConflict)
		return
	}
	err = client.DB("education_management").Get(request.Email, &existingDoc, couchdb.Options{})
	if err == nil {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create the faculty document
	doc := map[string]interface{}{
		"_id":         request.ID,
		"type":        request.TYPE,
		"name":        request.Name,
		"email":       request.Email,
		"designation": request.Designation,
		"password":    string(hashedPassword),
	}

	// Store the faculty document in the database
	_, err = client.DB("education_management").Put(request.ID, doc, "")
	if err != nil {
		http.Error(w, "failed to store faculty document", http.StatusInternalServerError)
		return
	}

	token := GenerateJWT(request.Email)
	response := map[string]string{
		"message": "Faculty registered successfully",
		"token":   token,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Faculty Login
func facultyLogin(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	var user struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := client.DB("education_management").Get(request.Email, &user, couchdb.Options{})
	if err != nil {
		http.Error(w, "email not found", http.StatusNotFound)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		http.Error(w, "incorrect password", http.StatusUnauthorized)
		return
	}

	token := GenerateJWT(request.Email)
	response := map[string]string{
		"message": "Login successful",
		"token":   token,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GenerateJWT generates a JWT token
func GenerateJWT(email string) string {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 7).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}

// JWT Middleware
func verifyJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "authorization header missing", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), emailKey, claims["email"])
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "invalid token", http.StatusUnauthorized)
		}
	})
}

func main() {
	// Initialize CouchDB client
	client, err := couchdb.NewClient("http://admin:admin@localhost:5984/", nil)
	if err != nil {
		log.Fatalf("Failed to connect to CouchDB: %v", err)
	}

	// Setup HTTP routes
	// http.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
	// 	switch r.Method {
	// 	case http.MethodPost:
	// 		student.CreateStudent(w, r, client) // Call the CreateStudent function
	// 	case http.MethodGet:
	// 		student.GetAllStudents(w, r, client) // Call GetAllStudents to retrieve all students
	// 	}
	// })

	// http.HandleFunc("/students/get", func(w http.ResponseWriter, r *http.Request) {
	// 	student.GetStudent(w, r, client) // Retrieve student by ID
	// })

	http.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			student.CreateStudent(w, r, client) // Call the CreateStudent function
		case http.MethodGet:
			// Secure the GetAllStudents endpoint with JWT verification
			verifyJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				student.GetAllStudents(w, r, client) // Pass the client to GetAllStudents
			})).ServeHTTP(w, r)
		}
	})

	http.HandleFunc("/students/get", func(w http.ResponseWriter, r *http.Request) {
		// Secure the GetStudent endpoint with JWT verification
		verifyJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			student.GetStudent(w, r, client) // Pass the client to GetStudent
		})).ServeHTTP(w, r)
	})
	// http.HandleFunc("/students/create", func(w http.ResponseWriter, r *http.Request) {
	// 	student.CreateStudent(w, r, client) // Retrieve student by ID
	// })

	// http.HandleFunc("/students/update", func(w http.ResponseWriter, r *http.Request) {
	// 	student.UpdateStudent(w, r, client) // Update student by ID
	// })

	// http.HandleFunc("/students/delete", func(w http.ResponseWriter, r *http.Request) {
	// 	student.DeleteStudent(w, r, client) // Delete student by ID
	// })

	http.HandleFunc("/students/create", func(w http.ResponseWriter, r *http.Request) {
		// Secure the CreateStudent endpoint with JWT verification
		verifyJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			student.CreateStudent(w, r, client) // Call CreateStudent function
		})).ServeHTTP(w, r)
	})

	http.HandleFunc("/students/update", func(w http.ResponseWriter, r *http.Request) {
		// Secure the UpdateStudent endpoint with JWT verification
		verifyJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			student.UpdateStudent(w, r, client) // Call UpdateStudent function
		})).ServeHTTP(w, r)
	})

	http.HandleFunc("/students/delete", func(w http.ResponseWriter, r *http.Request) {
		// Secure the DeleteStudent endpoint with JWT verification
		verifyJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			student.DeleteStudent(w, r, client) // Call DeleteStudent function
		})).ServeHTTP(w, r)
	})

	http.HandleFunc("/students/generate_qr", func(w http.ResponseWriter, r *http.Request) {
		student.GenerateAndSaveQRCode(w, r, client) // Generate QR code for a student
	})

	http.HandleFunc("/request-otp", requestOTP)
	http.HandleFunc("/register-faculty", func(w http.ResponseWriter, r *http.Request) {
		registerFaculty(w, r, client)
	})
	http.Handle("/faculty-login", verifyJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		facultyLogin(w, r, client)
	})))
	// Start the server
	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
