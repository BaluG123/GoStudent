package main

import (
	"log"
	"time"
)

// Common fields for all staff and students
type Person struct {
	ID               string    `json:"id"`
	FullName         string    `json:"full_name"`
	DateOfBirth      time.Time `json:"date_of_birth"`
	Gender           string    `json:"gender"`
	Address          string    `json:"address"`
	ContactNumber    string    `json:"contact_number"`
	EmailAddress     string    `json:"email_address"`
	EmergencyContact string    `json:"emergency_contact"`
}

// Student-specific fields
type Student struct {
	Person
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

// Teacher-specific fields
type Teacher struct {
	Person
	SubjectsTaught      []string             `json:"subjects_taught"`
	ClassAssigned       string               `json:"class_assigned"`
	Qualifications      []string             `json:"qualifications"`
	YearsOfExperience   int                  `json:"years_of_experience"`
	PreviousEmployment  []EmploymentRecord   `json:"previous_employment"`
	Certifications      []string             `json:"certifications"`
	AttendanceRecords   []AttendanceRecord   `json:"attendance_records"`
	PerformanceReviews  []PerformanceReview  `json:"performance_reviews"`
	Timetable           []TimetableEntry     `json:"timetable"`
	DateOfJoining       time.Time            `json:"date_of_joining"`
	SalaryDetails       Salary               `json:"salary_details"`
	LeaveRecords        []LeaveRecord        `json:"leave_records"`
	DisciplinaryRecords []DisciplinaryRecord `json:"disciplinary_records"`
}

// Cook staff-specific fields
type CookStaff struct {
	Person
	JobTitle            string               `json:"job_title"`
	Qualifications      []string             `json:"qualifications"`
	YearsOfExperience   int                  `json:"years_of_experience"`
	PreviousEmployment  []EmploymentRecord   `json:"previous_employment"`
	Certifications      []string             `json:"certifications"`
	AttendanceRecords   []AttendanceRecord   `json:"attendance_records"`
	PerformanceReviews  []PerformanceReview  `json:"performance_reviews"`
	DateOfJoining       time.Time            `json:"date_of_joining"`
	SalaryDetails       Salary               `json:"salary_details"`
	LeaveRecords        []LeaveRecord        `json:"leave_records"`
	DisciplinaryRecords []DisciplinaryRecord `json:"disciplinary_records"`
}

// Cleaning staff-specific fields
type CleaningStaff struct {
	Person
	JobTitle            string               `json:"job_title"`
	Qualifications      []string             `json:"qualifications"`
	YearsOfExperience   int                  `json:"years_of_experience"`
	PreviousEmployment  []EmploymentRecord   `json:"previous_employment"`
	Certifications      []string             `json:"certifications"`
	AttendanceRecords   []AttendanceRecord   `json:"attendance_records"`
	PerformanceReviews  []PerformanceReview  `json:"performance_reviews"`
	DateOfJoining       time.Time            `json:"date_of_joining"`
	SalaryDetails       Salary               `json:"salary_details"`
	LeaveRecords        []LeaveRecord        `json:"leave_records"`
	DisciplinaryRecords []DisciplinaryRecord `json:"disciplinary_records"`
}

// Supporting structures
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

type EmploymentRecord struct {
	Employer  string    `json:"employer"`
	JobTitle  string    `json:"job_title"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

type PerformanceReview struct {
	Date     time.Time `json:"date"`
	Reviewer string    `json:"reviewer"`
	Comments string    `json:"comments"`
	Rating   int       `json:"rating"`
}

type TimetableEntry struct {
	Day      string `json:"day"`
	TimeSlot string `json:"time_slot"`
	Subject  string `json:"subject"`
	Class    string `json:"class"`
}

type Salary struct {
	Amount  float64   `json:"amount"`
	PayDate time.Time `json:"pay_date"`
}

type LeaveRecord struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Reason    string    `json:"reason"`
}

type DisciplinaryRecord struct {
	Date        time.Time `json:"date"`
	Incident    string    `json:"incident"`
	ActionTaken string    `json:"action_taken"`
}

func main() {
	// Example usage
	student := Student{
		Person: Person{
			ID:               "1",
			FullName:         "John Doe",
			DateOfBirth:      time.Date(2005, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:           "Male",
			Address:          "123 Main St",
			ContactNumber:    "123-456-7890",
			EmailAddress:     "john.doe@example.com",
			EmergencyContact: "Jane Doe",
		},
		Class:            "10",
		Section:          "A",
		RollNumber:       "15",
		SubjectsEnrolled: []string{"Math", "Science", "English"},
		AttendanceRecords: []AttendanceRecord{
			{Date: time.Now(), Status: "Present"},
		},
		ExamScores: []ExamScore{
			{Subject: "Math", Score: 95, Grade: "A"},
		},
		ExtracurricularActivities: []string{"Basketball", "Debate Club"},
		BehavioralRecords: []BehavioralRecord{
			{Date: time.Now(), Incident: "Late to class", ActionTaken: "Warning"},
		},
		HealthRecords: []HealthRecord{
			{Condition: "Asthma", Notes: "Uses inhaler"},
		},
		AdmissionDate:  time.Now(),
		PreviousSchool: "ABC Elementary",
		FeePaymentRecords: []FeePaymentRecord{
			{Date: time.Now(), Amount: 500, Status: "Paid"},
		},
		Scholarships: []Scholarship{
			{Name: "Merit Scholarship", Amount: 1000, DateAwarded: time.Now()},
		},
	}

	log.Printf("Student: %+v", student)
}
