// user-service/services/user_service.go - Business Logic Layer
package services

import (
	"fmt"

	"gorm.io/gorm"

	"gitlab.com/nodiviti/user-service/database"
	"gitlab.com/nodiviti/user-service/models"
	"gitlab.com/nodiviti/user-service/utils"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService() *UserService {
	return &UserService{
		db: database.GetDB(),
	}
}

// GetUserByID retrieves user by ID
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	result := s.db.First(&user, id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByUsername retrieves user by username
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	result := s.db.Where("username = ? AND is_active = ?", username, true).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByEmail retrieves user by email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := s.db.Where("email = ? AND is_active = ?", email, true).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

// CreateUser creates a new user (removed - will be handled by auth-service register)
// This method is kept for admin-only user creation
func (s *UserService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	// Check if user already exists
	exists, err := s.CheckUserExists(req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username or email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Create user model
	user := models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
		IsActive:     true,

		// Profile fields (optional)
		FullName:    req.FullName,
		Phone:       req.Phone,
		Address:     req.Address,
		DateOfBirth: req.DateOfBirth,
		Gender:      req.Gender,

		// Role-specific fields (optional)
		EmployeeID:     req.EmployeeID,
		StudentID:      req.StudentID,
		ClassLevel:     req.ClassLevel,
		AcademicYear:   req.AcademicYear,
		ParentName:     req.ParentName,
		ParentPhone:    req.ParentPhone,
		Specialization: req.Specialization,
	}

	// Create user in database with GORM
	result := s.db.Create(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %v", result.Error)
	}

	return &user, nil
}

// UpdateUser updates user profile
func (s *UserService) UpdateUser(userID uint, req *models.UpdateUserRequest) (*models.User, error) {
	var user models.User
	result := s.db.First(&user, userID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error
	}

	// Update fields if provided
	updateData := make(map[string]interface{})

	if req.FullName != nil {
		updateData["full_name"] = req.FullName
	}
	if req.Phone != nil {
		updateData["phone"] = req.Phone
	}
	if req.Address != nil {
		updateData["address"] = req.Address
	}
	if req.DateOfBirth != nil {
		updateData["date_of_birth"] = req.DateOfBirth
	}
	if req.Gender != nil {
		updateData["gender"] = req.Gender
	}
	if req.EmployeeID != nil {
		updateData["employee_id"] = req.EmployeeID
	}
	if req.StudentID != nil {
		updateData["student_id"] = req.StudentID
	}
	if req.ClassLevel != nil {
		updateData["class_level"] = req.ClassLevel
	}
	if req.AcademicYear != nil {
		updateData["academic_year"] = req.AcademicYear
	}
	if req.ParentName != nil {
		updateData["parent_name"] = req.ParentName
	}
	if req.ParentPhone != nil {
		updateData["parent_phone"] = req.ParentPhone
	}
	if req.Specialization != nil {
		updateData["specialization"] = req.Specialization
	}
	if req.ExperienceYears != nil {
		updateData["experience_years"] = req.ExperienceYears
	}
	if req.EmergencyContact != nil {
		updateData["emergency_contact"] = req.EmergencyContact
	}
	if req.EmergencyPhone != nil {
		updateData["emergency_phone"] = req.EmergencyPhone
	}
	if req.MedicalConditions != nil {
		updateData["medical_conditions"] = req.MedicalConditions
	}
	if req.Status != nil {
		updateData["status"] = req.Status
	}

	// Update user
	result = s.db.Model(&user).Updates(updateData)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update user: %v", result.Error)
	}

	// Fetch updated user
	s.db.First(&user, userID)

	return &user, nil
}

// GetAllUsers retrieves users with pagination and filters
func (s *UserService) GetAllUsers(page, limit int, role string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := s.db.Model(&models.User{}).Where("is_active = ?", true)

	// Apply role filter if specified
	if role != "" {
		query = query.Where("role = ?", role)
	}

	// Get total count
	countResult := query.Count(&total)
	if countResult.Error != nil {
		return nil, 0, countResult.Error
	}

	// Get paginated data
	offset := (page - 1) * limit
	result := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return users, total, nil
}

// GetUsersByRole retrieves users by role
func (s *UserService) GetUsersByRole(role string) ([]models.User, error) {
	var users []models.User
	result := s.db.Where("role = ? AND is_active = ?", role, true).Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

// GetTeachers retrieves all teachers with their specialization
func (s *UserService) GetTeachers() ([]models.User, error) {
	var teachers []models.User
	result := s.db.Where("role = ? AND is_active = ?", "teacher", true).
		Where("employee_id IS NOT NULL AND specialization IS NOT NULL").
		Find(&teachers)

	if result.Error != nil {
		return nil, result.Error
	}

	return teachers, nil
}

// GetStudents retrieves all students with class info
func (s *UserService) GetStudents() ([]models.User, error) {
	var students []models.User
	result := s.db.Where("role = ? AND is_active = ?", "student", true).
		Where("student_id IS NOT NULL AND class_level IS NOT NULL").
		Find(&students)

	if result.Error != nil {
		return nil, result.Error
	}

	return students, nil
}

// GetStudentsByClass retrieves students by class level
func (s *UserService) GetStudentsByClass(classLevel string) ([]models.User, error) {
	var students []models.User
	result := s.db.Where("role = ? AND class_level = ? AND is_active = ?", "student", classLevel, true).
		Find(&students)

	if result.Error != nil {
		return nil, result.Error
	}

	return students, nil
}

// UpdateUserPhoto updates user profile photo
func (s *UserService) UpdateUserPhoto(userID uint, photoPath string) error {
	result := s.db.Model(&models.User{}).Where("id = ?", userID).Update("profile_photo", photoPath)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeactivateUser soft deletes user
func (s *UserService) DeactivateUser(userID uint) error {
	result := s.db.Model(&models.User{}).Where("id = ?", userID).Update("is_active", false)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// ActivateUser reactivates user
func (s *UserService) ActivateUser(userID uint) error {
	result := s.db.Model(&models.User{}).Where("id = ?", userID).Update("is_active", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeleteUser permanently deletes user (GORM soft delete)
func (s *UserService) DeleteUser(userID uint) error {
	result := s.db.Delete(&models.User{}, userID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// CheckUserExists checks if username or email already exists
func (s *UserService) CheckUserExists(username, email string) (bool, error) {
	var count int64
	result := s.db.Model(&models.User{}).Where("username = ? OR email = ?", username, email).Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

// GetUserStats returns user statistics
func (s *UserService) GetUserStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total active users
	var totalActive int64
	s.db.Model(&models.User{}).Where("is_active = ?", true).Count(&totalActive)
	stats["total_active"] = totalActive

	// Users by role
	var admins int64
	s.db.Model(&models.User{}).Where("role = ? AND is_active = ?", "admin", true).Count(&admins)
	stats["admins"] = admins

	var teachers int64
	s.db.Model(&models.User{}).Where("role = ? AND is_active = ?", "teacher", true).Count(&teachers)
	stats["teachers"] = teachers

	var students int64
	s.db.Model(&models.User{}).Where("role = ? AND is_active = ?", "student", true).Count(&students)
	stats["students"] = students

	// Inactive users
	var inactive int64
	s.db.Model(&models.User{}).Where("is_active = ?", false).Count(&inactive)
	stats["inactive"] = inactive

	return stats, nil
}

// SearchUsers searches users by name, username, or email
func (s *UserService) SearchUsers(query string, role string, limit int) ([]models.User, error) {
	var users []models.User

	db := s.db.Where("is_active = ?", true)

	if role != "" {
		db = db.Where("role = ?", role)
	}

	searchPattern := "%" + query + "%"
	db = db.Where(
		s.db.Where("full_name ILIKE ?", searchPattern).
			Or("username ILIKE ?", searchPattern).
			Or("email ILIKE ?", searchPattern),
	)

	result := db.Limit(limit).Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}

// GetUserWithProfile gets user with complete profile based on role
func (s *UserService) GetUserWithProfile(userID uint) (*models.User, error) {
	var user models.User
	result := s.db.First(&user, userID)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, result.Error
	}

	return &user, nil
}

// ValidateRoleRequiredFields validates that role-specific required fields are present
func (s *UserService) ValidateRoleRequiredFields(user *models.User) error {
	switch user.Role {
	case "teacher":
		if user.EmployeeID == nil || *user.EmployeeID == "" {
			return fmt.Errorf("employee_id is required for teachers")
		}
		if user.Specialization == nil || *user.Specialization == "" {
			return fmt.Errorf("specialization is required for teachers")
		}
	case "student":
		if user.StudentID == nil || *user.StudentID == "" {
			return fmt.Errorf("student_id is required for students")
		}
		if user.ClassLevel == nil || *user.ClassLevel == "" {
			return fmt.Errorf("class_level is required for students")
		}
		if user.ParentName == nil || *user.ParentName == "" {
			return fmt.Errorf("parent_name is required for students")
		}
		if user.ParentPhone == nil || *user.ParentPhone == "" {
			return fmt.Errorf("parent_phone is required for students")
		}
	case "admin":
		// Admin doesn't require specific fields, but employee_id is recommended
	}
	return nil
}

// BulkCreateUsers creates multiple users (useful for imports)
func (s *UserService) BulkCreateUsers(users []models.User) error {
	// Validate all users first
	for i, user := range users {
		if err := s.ValidateRoleRequiredFields(&user); err != nil {
			return fmt.Errorf("validation failed for user %d: %v", i+1, err)
		}
	}

	// Create all users in a transaction
	result := s.db.CreateInBatches(&users, 100) // Process in batches of 100

	if result.Error != nil {
		return fmt.Errorf("bulk create failed: %v", result.Error)
	}

	return nil
}

// GetClassList returns list of all classes
func (s *UserService) GetClassList() ([]string, error) {
	var classes []string
	result := s.db.Model(&models.User{}).
		Where("role = ? AND is_active = ? AND class_level IS NOT NULL", "student", true).
		Distinct("class_level").
		Pluck("class_level", &classes)

	if result.Error != nil {
		return nil, result.Error
	}

	return classes, nil
}

// GetSpecializationList returns list of all teacher specializations
func (s *UserService) GetSpecializationList() ([]string, error) {
	var specializations []string
	result := s.db.Model(&models.User{}).
		Where("role = ? AND is_active = ? AND specialization IS NOT NULL", "teacher", true).
		Distinct("specialization").
		Pluck("specialization", &specializations)

	if result.Error != nil {
		return nil, result.Error
	}

	return specializations, nil
}
