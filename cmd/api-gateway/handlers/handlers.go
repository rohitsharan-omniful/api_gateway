package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/go-service-kit/pkg/errors"
	"github.com/yourorg/go-service-kit/pkg/httpservice"
)

// GetUserHandler demonstrates a simple GET handler with zero boilerplate.
// Notice: No manual logging, no logger retrieval, just business logic.
func GetUserHandler(c *gin.Context) error {
	userID := c.Param("id")

	// Simulate fetching user
	user := map[string]interface{}{
		"id":    userID,
		"name":  "John Doe",
		"email": "john@example.com",
	}

	httpservice.RespondSuccess(c, user)
	return nil
}

// CreateOrderHandler demonstrates a POST handler with validation.
func CreateOrderHandler(c *gin.Context) error {
	var req struct {
		ProductID string  `json:"product_id" binding:"required"`
		Quantity  int     `json:"quantity" binding:"required,min=1"`
		Price     float64 `json:"price" binding:"required,min=0"`
	}

	// Validate request
	if !httpservice.ValidateJSON(c, &req) {
		return nil // Validation already handled the response
	}

	// Simulate order creation
	order := map[string]interface{}{
		"order_id":   "ORD-12345",
		"product_id": req.ProductID,
		"quantity":   req.Quantity,
		"total":      req.Price * float64(req.Quantity),
		"status":     "pending",
	}

	httpservice.RespondCreated(c, order)
	return nil
}

// ErrorExampleHandler demonstrates error handling with zero boilerplate.
// Just return the error - the wrapper handles logging and response.
func ErrorExampleHandler(c *gin.Context) error {
	// Simulate a business logic error
	shouldFail := c.Query("fail")

	if shouldFail == "true" {
		// Just return the error - wrapper handles everything
		return errors.NewNotFoundError("Resource not found")
	}

	httpservice.RespondSuccess(c, gin.H{
		"message": "Success! Try adding ?fail=true to see error handling",
	})
	return nil
}

// ComplexBusinessLogicHandler demonstrates a handler with multiple steps.
// Notice: Still zero logging boilerplate, just business logic.
func ComplexBusinessLogicHandler(c *gin.Context) error {
	// Step 1: Validate input
	itemID := c.Param("id")
	if itemID == "" {
		return errors.NewValidationError("Item ID is required")
	}

	// Step 2: Simulate fetching from database
	// In real code, this might return an error
	item, err := fetchItemFromDB(itemID)
	if err != nil {
		return err // Wrapper logs and handles this automatically
	}

	// Step 3: Simulate business logic
	processedItem := processItem(item)

	// Step 4: Return success
	httpservice.RespondSuccess(c, processedItem)
	return nil
}

// Helper functions (simulate business logic)
func fetchItemFromDB(id string) (map[string]interface{}, error) {
	if id == "error" {
		return nil, errors.NewInternalError("Database connection failed")
	}
	return map[string]interface{}{
		"id":   id,
		"name": "Sample Item",
	}, nil
}

func processItem(item map[string]interface{}) map[string]interface{} {
	item["processed"] = true
	item["timestamp"] = "2025-11-21T15:00:00Z"
	return item
}

// CreateUserHandler creates a new user - demonstrates zero-boilerplate pattern.
// Notice: No logging code at all, just business logic!
func CreateUserHandler(c *gin.Context) error {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	// Validate request - this handles response if validation fails
	if !httpservice.ValidateJSON(c, &req) {
		return nil
	}

	// Simulate user creation (in real app, this would call a service/repository)
	user := map[string]interface{}{
		"id":     "USR-" + generateID(),
		"name":   req.Name,
		"email":  req.Email,
		"status": "active",
	}

	// Return success - wrapper logs everything automatically
	httpservice.RespondCreated(c, user)
	return nil
}

// UpdateUserHandler updates an existing user.
func UpdateUserHandler(c *gin.Context) error {
	userID := c.Param("id")

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email" binding:"omitempty,email"`
	}

	if !httpservice.ValidateJSON(c, &req) {
		return nil
	}

	// Simulate checking if user exists
	if userID == "not-found" {
		return errors.NewNotFoundError(fmt.Sprintf("User %s not found", userID))
	}

	// Simulate update
	updatedUser := map[string]interface{}{
		"id":    userID,
		"name":  req.Name,
		"email": req.Email,
	}

	httpservice.RespondSuccess(c, updatedUser)
	return nil
}

// DeleteUserHandler deletes a user.
func DeleteUserHandler(c *gin.Context) error {
	userID := c.Param("id")

	// Simulate checking if user exists
	if userID == "not-found" {
		return errors.NewNotFoundError(fmt.Sprintf("User %s not found", userID))
	}

	// Simulate deletion
	httpservice.RespondSuccess(c, gin.H{
		"message": "User deleted successfully",
		"id":      userID,
	})
	return nil
}

// Helper function
func generateID() string {
	return "12345" // In real app, use UUID
}
