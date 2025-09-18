package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"gitlab.com/nodiviti/user-service/config"
	"gitlab.com/nodiviti/user-service/models"
	"gitlab.com/nodiviti/user-service/services"
	"gitlab.com/nodiviti/user-service/utils"
)

type UserHandler struct {
	cfg         *config.Config
	validator   *validator.Validate
	userService *services.UserService
}

func NewUserHandler(cfg *config.Config, userService *services.UserService) *UserHandler {
	return &UserHandler{
		cfg:         cfg,
		validator:   validator.New(),
		userService: userService,
	}
}

// GetMyProfile retrieves current user's complete profile
func (h *UserHandler) GetMyProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	// Convert to uint (GORM uses uint for ID)
	id := uint(userID.(int))

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile retrieved successfully",
		"data":    user.ToResponse(), // Remove sensitive fields
	})
}

// UpdateMyProfile updates current user's profile
func (h *UserHandler) UpdateMyProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	id := uint(userID.(int))
	user, err := h.userService.UpdateUser(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"data":    user.ToResponse(),
	})
}

// GetUserByID retrieves user profile by ID (admin/teacher access)
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User profile retrieved successfully",
		"data":    user.ToResponse(),
	})
}

// GetAllUsers retrieves all users with pagination (admin only)
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	role := c.Query("role") // Optional role filter

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := h.userService.GetAllUsers(page, limit, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve users",
		})
		return
	}

	// Convert to response format
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, *user.ToResponse())
	}

	// Calculate pagination info
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, gin.H{
		"message": "Users retrieved successfully",
		"data":    userResponses,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
		"filters": gin.H{
			"role": role,
		},
	})
}

// CreateUser creates a new user (admin only)
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.CreateUser(&req)
	if err != nil {
		if err.Error() == "username or email already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create user",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"data":    user.ToResponse(),
	})
}

// UpdateUser updates user profile (admin only)
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	user, err := h.userService.UpdateUser(uint(userID), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"data":    user.ToResponse(),
	})
}

// DeactivateUser deactivates a user account (admin only)
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	err = h.userService.DeactivateUser(uint(userID))
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to deactivate user",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deactivated successfully",
	})
}

// GetTeachers retrieves all teachers
func (h *UserHandler) GetTeachers(c *gin.Context) {
	teachers, err := h.userService.GetTeachers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve teachers",
		})
		return
	}

	var teacherResponses []models.UserResponse
	for _, teacher := range teachers {
		teacherResponses = append(teacherResponses, *teacher.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Teachers retrieved successfully",
		"data":    teacherResponses,
		"count":   len(teacherResponses),
	})
}

// GetStudents retrieves all students
func (h *UserHandler) GetStudents(c *gin.Context) {
	students, err := h.userService.GetStudents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve students",
		})
		return
	}

	var studentResponses []models.UserResponse
	for _, student := range students {
		studentResponses = append(studentResponses, *student.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Students retrieved successfully",
		"data":    studentResponses,
		"count":   len(studentResponses),
	})
}

// GetStudentsByClass retrieves students by class level
func (h *UserHandler) GetStudentsByClass(c *gin.Context) {
	classLevel := c.Param("class")
	if classLevel == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Class level is required",
		})
		return
	}

	students, err := h.userService.GetStudentsByClass(classLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve students",
		})
		return
	}

	var studentResponses []models.UserResponse
	for _, student := range students {
		studentResponses = append(studentResponses, *student.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Students retrieved successfully",
		"data":    studentResponses,
		"class":   classLevel,
		"count":   len(studentResponses),
	})
}

// GetClassList retrieves list of all classes
func (h *UserHandler) GetClassList(c *gin.Context) {
	classes, err := h.userService.GetClassList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve class list",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Class list retrieved successfully",
		"data":    classes,
		"count":   len(classes),
	})
}

// GetUserStats returns user statistics (admin only)
func (h *UserHandler) GetUserStats(c *gin.Context) {
	stats, err := h.userService.GetUserStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User statistics retrieved successfully",
		"data":    stats,
	})
}

// SearchUsers searches users by query (admin only)
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	role := c.Query("role")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 100
	}

	users, err := h.userService.SearchUsers(query, role, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search users",
		})
		return
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, *user.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Search completed successfully",
		"data":    userResponses,
		"query":   query,
		"role":    role,
		"count":   len(userResponses),
	})
}

// UploadProfilePhoto handles profile photo upload
func (h *UserHandler) UploadProfilePhoto(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in context",
		})
		return
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file uploaded",
		})
		return
	}

	// Validate file
	if err := utils.ValidateImageFile(file, h.cfg.Upload.MaxSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Save file
	filename, err := utils.SaveUploadedFile(file, "profiles", userID.(int), h.cfg.Upload.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file",
		})
		return
	}

	// Update profile photo path in database
	id := uint(userID.(int))
	err = h.userService.UpdateUserPhoto(id, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update profile photo",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile photo updated successfully",
		"photo":   filename,
		"url":     "/files/" + filename,
	})
}

// HealthCheck returns service health status
func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"service":  h.cfg.ServiceName,
		"version":  h.cfg.Version,
		"database": "gorm+postgresql",
	})
}
