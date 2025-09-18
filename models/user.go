// user-service/models/user.go - Simple Single Table with GORM
package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// User represents complete user data in single table
type User struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // Soft delete

	// Auth fields (required)
	Username     string `json:"username" gorm:"uniqueIndex;size:100;not null"`
	Email        string `json:"email" gorm:"uniqueIndex;size:255;not null"`
	PasswordHash string `json:"-" gorm:"size:255;not null"`
	Role         string `json:"role" gorm:"size:20;not null;check:role IN ('admin','teacher','student')"`
	IsActive     bool   `json:"is_active" gorm:"default:true;index"`

	// Basic Profile fields (all optional)
	FullName     *string    `json:"full_name,omitempty" gorm:"size:255"`
	Phone        *string    `json:"phone,omitempty" gorm:"size:20"`
	Address      *string    `json:"address,omitempty" gorm:"type:text"`
	DateOfBirth  *time.Time `json:"date_of_birth,omitempty" gorm:"type:date"`
	Gender       *string    `json:"gender,omitempty" gorm:"size:10;check:gender IN ('male','female')"`
	ProfilePhoto *string    `json:"profile_photo,omitempty" gorm:"size:500"`

	// Role-specific fields (optional, depends on role)
	// Teacher fields
	EmployeeID      *string    `json:"employee_id,omitempty" gorm:"uniqueIndex;size:50"` // For teachers & admins
	Specialization  *string    `json:"specialization,omitempty" gorm:"size:255"`         // For teachers
	Qualification   *string    `json:"qualification,omitempty" gorm:"type:text"`         // For teachers
	ExperienceYears *int       `json:"experience_years,omitempty" gorm:"default:0"`      // For teachers
	HireDate        *time.Time `json:"hire_date,omitempty" gorm:"type:date"`             // For teachers & admins
	Salary          *float64   `json:"salary,omitempty" gorm:"type:decimal(12,2)"`       // For teachers

	// Student fields
	StudentID      *string    `json:"student_id,omitempty" gorm:"uniqueIndex;size:50"` // For students
	ClassLevel     *string    `json:"class_level,omitempty" gorm:"size:50"`            // For students
	AcademicYear   *string    `json:"academic_year,omitempty" gorm:"size:20"`          // For students
	ParentName     *string    `json:"parent_name,omitempty" gorm:"size:255"`           // For students
	ParentPhone    *string    `json:"parent_phone,omitempty" gorm:"size:20"`           // For students
	ParentEmail    *string    `json:"parent_email,omitempty" gorm:"size:255"`          // For students
	EnrollmentDate *time.Time `json:"enrollment_date,omitempty" gorm:"type:date"`      // For students
	GraduationDate *time.Time `json:"graduation_date,omitempty" gorm:"type:date"`      // For students

	// Optional fields for all roles
	EmergencyContact  *string `json:"emergency_contact,omitempty" gorm:"size:255"`
	EmergencyPhone    *string `json:"emergency_phone,omitempty" gorm:"size:20"`
	MedicalConditions *string `json:"medical_conditions,omitempty" gorm:"type:text"`
	BloodType         *string `json:"blood_type,omitempty" gorm:"size:5"`

	// Status field (role-specific meaning)
	Status *string `json:"status,omitempty" gorm:"size:20;default:'active'"` // active, inactive, graduated, etc

	// JSON field for additional flexible data
	AdditionalData *string `json:"additional_data,omitempty" gorm:"type:jsonb"` // PostgreSQL JSONB
}

// Request/Response DTOs
type CreateUserRequest struct {
	// Auth data (required)
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"required,oneof=admin teacher student"`

	// Profile data (optional)
	FullName    *string    `json:"full_name,omitempty" validate:"omitempty,min=2,max=255"`
	Phone       *string    `json:"phone,omitempty" validate:"omitempty,min=10,max=20"`
	Address     *string    `json:"address,omitempty" validate:"omitempty,max=1000"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Gender      *string    `json:"gender,omitempty" validate:"omitempty,oneof=male female"`

	// Role-specific fields
	EmployeeID     *string `json:"employee_id,omitempty"`
	StudentID      *string `json:"student_id,omitempty"`
	ClassLevel     *string `json:"class_level,omitempty"`
	AcademicYear   *string `json:"academic_year,omitempty"`
	ParentName     *string `json:"parent_name,omitempty"`
	ParentPhone    *string `json:"parent_phone,omitempty"`
	Specialization *string `json:"specialization,omitempty"`
}

type UpdateUserRequest struct {
	// Profile fields (all optional)
	FullName    *string    `json:"full_name,omitempty" validate:"omitempty,min=2,max=255"`
	Phone       *string    `json:"phone,omitempty" validate:"omitempty,min=10,max=20"`
	Address     *string    `json:"address,omitempty" validate:"omitempty,max=1000"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Gender      *string    `json:"gender,omitempty" validate:"omitempty,oneof=male female"`

	// Role-specific updates
	EmployeeID      *string `json:"employee_id,omitempty"`
	StudentID       *string `json:"student_id,omitempty"`
	ClassLevel      *string `json:"class_level,omitempty"`
	AcademicYear    *string `json:"academic_year,omitempty"`
	ParentName      *string `json:"parent_name,omitempty"`
	ParentPhone     *string `json:"parent_phone,omitempty"`
	Specialization  *string `json:"specialization,omitempty"`
	ExperienceYears *int    `json:"experience_years,omitempty"`

	// Optional fields
	EmergencyContact  *string `json:"emergency_contact,omitempty"`
	EmergencyPhone    *string `json:"emergency_phone,omitempty"`
	MedicalConditions *string `json:"medical_conditions,omitempty"`
	Status            *string `json:"status,omitempty"`
}

// UserResponse for API responses (without sensitive data)
type UserResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`

	// Profile data
	FullName     *string    `json:"full_name,omitempty"`
	Phone        *string    `json:"phone,omitempty"`
	Address      *string    `json:"address,omitempty"`
	DateOfBirth  *time.Time `json:"date_of_birth,omitempty"`
	Gender       *string    `json:"gender,omitempty"`
	ProfilePhoto *string    `json:"profile_photo,omitempty"`

	// Role-specific data (based on role)
	EmployeeID      *string    `json:"employee_id,omitempty"`
	Specialization  *string    `json:"specialization,omitempty"`
	Qualification   *string    `json:"qualification,omitempty"`
	ExperienceYears *int       `json:"experience_years,omitempty"`
	HireDate        *time.Time `json:"hire_date,omitempty"`

	StudentID      *string    `json:"student_id,omitempty"`
	ClassLevel     *string    `json:"class_level,omitempty"`
	AcademicYear   *string    `json:"academic_year,omitempty"`
	ParentName     *string    `json:"parent_name,omitempty"`
	ParentPhone    *string    `json:"parent_phone,omitempty"`
	ParentEmail    *string    `json:"parent_email,omitempty"`
	EnrollmentDate *time.Time `json:"enrollment_date,omitempty"`
	GraduationDate *time.Time `json:"graduation_date,omitempty"`

	EmergencyContact  *string `json:"emergency_contact,omitempty"`
	EmergencyPhone    *string `json:"emergency_phone,omitempty"`
	MedicalConditions *string `json:"medical_conditions,omitempty"`
	Status            *string `json:"status,omitempty"`
}

// Convert User to UserResponse (remove sensitive fields)
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		IsActive:  u.IsActive,

		FullName:     u.FullName,
		Phone:        u.Phone,
		Address:      u.Address,
		DateOfBirth:  u.DateOfBirth,
		Gender:       u.Gender,
		ProfilePhoto: u.ProfilePhoto,

		EmployeeID:      u.EmployeeID,
		Specialization:  u.Specialization,
		Qualification:   u.Qualification,
		ExperienceYears: u.ExperienceYears,
		HireDate:        u.HireDate,

		StudentID:      u.StudentID,
		ClassLevel:     u.ClassLevel,
		AcademicYear:   u.AcademicYear,
		ParentName:     u.ParentName,
		ParentPhone:    u.ParentPhone,
		ParentEmail:    u.ParentEmail,
		EnrollmentDate: u.EnrollmentDate,
		GraduationDate: u.GraduationDate,

		EmergencyContact:  u.EmergencyContact,
		EmergencyPhone:    u.EmergencyPhone,
		MedicalConditions: u.MedicalConditions,
		Status:            u.Status,
	}
}

// Validation methods
func (u *User) ValidateForRole() error {
	switch u.Role {
	case "teacher":
		if u.EmployeeID == nil || *u.EmployeeID == "" {
			return fmt.Errorf("employee_id is required for teachers")
		}
		if u.Specialization == nil || *u.Specialization == "" {
			return fmt.Errorf("specialization is required for teachers")
		}
	case "student":
		if u.StudentID == nil || *u.StudentID == "" {
			return fmt.Errorf("student_id is required for students")
		}
		if u.ClassLevel == nil || *u.ClassLevel == "" {
			return fmt.Errorf("class_level is required for students")
		}
		if u.ParentName == nil || *u.ParentName == "" {
			return fmt.Errorf("parent_name is required for students")
		}
		if u.ParentPhone == nil || *u.ParentPhone == "" {
			return fmt.Errorf("parent_phone is required for students")
		}
	case "admin":
		// Admin might need employee_id but it's optional
	}
	return nil
}

// GORM Hooks
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Validate role-specific required fields
	return u.ValidateForRole()
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Only validate if role is being changed
	if tx.Statement.Changed("Role") {
		return u.ValidateForRole()
	}
	return nil
}
