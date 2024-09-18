package staff

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
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

type SchoolStaff struct {
	ID                      string      `json:"id"`
	FullName                string      `json:"full_name"`
	DateOfBirth             CustomTime  `json:"date_of_birth"`
	Gender                  string      `json:"gender"`
	ContactNumber           string      `json:"contactNumber"`
	EmailAddress            string      `json:"emailAddress"`
	EmergencyContact        string      `json:"emergencyContact"`
	JobTitle                string      `json:"jobTitle"`
	Department              string      `json:"department"`
	StartDate               CustomTime  `json:"startDate"`
	Salary                  float64     `json:"salary"`
	Benefits                []string    `json:"benefits"`
	EducationLevel          string      `json:"educationLevel"`
	Certifications          []string    `json:"certifications"`
	Experience              int         `json:"experience"`
	ProfessionalDevelopment []string    `json:"professionalDevelopment"`
	CEUs                    int         `json:"CEUs"`
	EmployeeID              string      `json:"employeeID"`
	EmploymentStatus        string      `json:"employmentStatus"`
	WorkHours               string      `json:"workHours"`
	TimeOff                 []TimeOff   `json:"timeOff"`
	PayrollInfo             PayrollInfo `json:"payrollInfo"`
}

type TimeOff struct {
	Type  string     `json:"type"`
	Hours int        `json:"hours"`
	Date  CustomTime `json:"date"`
}

type PayrollInfo struct {
	DirectDeposit   string             `json:"directDeposit"`
	TaxWithholdings map[string]float64 `json:"taxWithholdings"`
}

func CreateStaff(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var staff SchoolStaff

	// Debug: Log the incoming request body
	log.Println("Received request to create staff")

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// Debug: Log the raw request body
	log.Printf("Raw request body: %s", string(body))

	// Decode the request body into the staff struct
	if err := json.Unmarshal(body, &staff); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	// Debug: Log the decoded staff struct
	log.Printf("Decoded staff: %+v", staff)

	var existingDoc map[string]interface{}
	err = client.DB("staff_db").Get(staff.ID, &existingDoc, couchdb.Options{})
	if err == nil {
		http.Error(w, "Staff ID already exists", http.StatusConflict)
		return
	}

	doc := map[string]interface{}{
		"_id":                      staff.ID,
		"full_name":                staff.FullName,
		"date_of_birth":            staff.DateOfBirth.Format(ctLayout), // Format date for CouchDB
		"gender":                   staff.Gender,
		"contact_number":           staff.ContactNumber,
		"email_address":            staff.EmailAddress,
		"emergency_contact":        staff.EmergencyContact,
		"job_title":                staff.JobTitle,
		"department":               staff.Department,
		"start_date":               staff.StartDate.Format(ctLayout), // Format date for CouchDB
		"salary":                   staff.Salary,
		"benefits":                 staff.Benefits,
		"education_level":          staff.EducationLevel,
		"certifications":           staff.Certifications,
		"experience":               staff.Experience,
		"professional_development": staff.ProfessionalDevelopment,
		"CEUs":                     staff.CEUs,
		"employee_id":              staff.EmployeeID,
		"employment_status":        staff.EmploymentStatus,
		"work_hours":               staff.WorkHours,
		"time_off":                 staff.TimeOff,
		"payroll_info":             staff.PayrollInfo,
	}

	_, err = client.DB("staff_db").Put(staff.ID, doc, "")
	if err != nil {
		http.Error(w, "failed to create staff", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Staff member created successfully"})
}

// Retrieve a staff member by ID
func GetStaff(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	staffID := r.URL.Query().Get("id")
	if staffID == "" {
		http.Error(w, "Staff ID missing", http.StatusBadRequest)
		return
	}

	var staff map[string]interface{}
	err := client.DB("staff_db").Get(staffID, &staff, couchdb.Options{})
	if err != nil {
		http.Error(w, "Staff member not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(staff)
}

func GetAllStaff(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	// Define a struct for the result of AllDocs
	type AllDocsResult struct {
		Rows []struct {
			ID  string          `json:"id"`
			Doc json.RawMessage `json:"doc"`
		} `json:"rows"`
	}

	var result AllDocsResult

	// Fetch all documents (students) from the "student_db"
	err := client.DB("staff_db").AllDocs(&result, couchdb.Options{
		"include_docs": true, // Include full documents, not just IDs
	})
	if err != nil {
		http.Error(w, "Failed to fetch staff members", http.StatusInternalServerError)
		return
	}

	// Prepare a slice to store the staff details
	var staffs []map[string]interface{}

	// Iterate over the rows and append the staff details to the slice
	for _, row := range result.Rows {
		var staff map[string]interface{}
		if err := json.Unmarshal(row.Doc, &staff); err == nil {
			// Exclude the _rev field if necessary
			delete(staff, "_rev")
			staffs = append(staffs, staff)
		}
	}

	// Respond with the list of staff members
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(staffs)
}

func UpdateStaff(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	var staff SchoolStaff

	if err := json.NewDecoder(r.Body).Decode(&staff); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	var existingDoc map[string]interface{}
	err := client.DB("staff_db").Get(staff.ID, &existingDoc, couchdb.Options{})
	if err != nil {
		http.Error(w, "Staff member not found", http.StatusNotFound)
		return
	}

	doc := map[string]interface{}{
		"_id":                      staff.ID,
		"_rev":                     existingDoc["_rev"],
		"full_name":                staff.FullName,
		"date_of_birth":            staff.DateOfBirth.Format(ctLayout), // Format date for CouchDB
		"gender":                   staff.Gender,
		"contact_number":           staff.ContactNumber,
		"email_address":            staff.EmailAddress,
		"emergency_contact":        staff.EmergencyContact,
		"job_title":                staff.JobTitle,
		"department":               staff.Department,
		"start_date":               staff.StartDate.Format(ctLayout), // Format date for CouchDB
		"salary":                   staff.Salary,
		"benefits":                 staff.Benefits,
		"education_level":          staff.EducationLevel,
		"certifications":           staff.Certifications,
		"experience":               staff.Experience,
		"professional_development": staff.ProfessionalDevelopment,
		"CEUs":                     staff.CEUs,
		"employee_id":              staff.EmployeeID,
		"employment_status":        staff.EmploymentStatus,
		"work_hours":               staff.WorkHours,
		"time_off":                 staff.TimeOff,
		"payroll_info":             staff.PayrollInfo,
	}

	_, err = client.DB("staff_db").Put(staff.ID, doc, existingDoc["_rev"].(string))
	if err != nil {
		http.Error(w, "failed to update staff member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Staff member updated successfully"})
}

func DeleteStaff(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	staffID := r.URL.Query().Get("id")
	if staffID == "" {
		http.Error(w, "Staff ID missing", http.StatusBadRequest)
		return
	}

	var existingDoc map[string]interface{}
	err := client.DB("staff_db").Get(staffID, &existingDoc, couchdb.Options{})
	if err != nil {
		http.Error(w, "Staff member not found", http.StatusNotFound)
		return
	}

	_, err = client.DB("staff_db").Delete(staffID, existingDoc["_rev"].(string))
	if err != nil {
		http.Error(w, "failed to delete staff member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Staff member deleted successfully"})
}

func GenerateAndSaveStaffQRCode(w http.ResponseWriter, r *http.Request, client *couchdb.Client) {
	staffID := r.URL.Query().Get("id")
	if staffID == "" {
		http.Error(w, "Staff ID missing", http.StatusBadRequest)
		return
	}

	// Fetch staff data from CouchDB
	var staff map[string]interface{}
	err := client.DB("staff_db").Get(staffID, &staff, couchdb.Options{})
	if err != nil {
		http.Error(w, "Staff member not found", http.StatusNotFound)
		return
	}

	// Create a map to store the data for the QR code
	staffDataForQR := map[string]interface{}{
		"id":             staff["_id"],
		"full_name":      staff["full_name"],
		"contact_number": staff["contact_number"],
		"email_address":  staff["email_address"],
		"job_title":      staff["jobtitle"],
	}

	// Convert the QR code data to JSON
	qrData, err := json.Marshal(staffDataForQR)
	if err != nil {
		http.Error(w, "Failed to marshal staff data for QR code", http.StatusInternalServerError)
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

	// Add the Base64 QR code to the staff document
	staff["qr_code"] = qrCodeBase64

	// Update the staff document in the database
	_, err = client.DB("staff_db").Put(staffID, staff, staff["_rev"].(string))
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
