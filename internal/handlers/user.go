package handlers

import (
	"strconv"
	"strings"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/httpclient"
	"golang-boilerplate/internal/integration/auth"
	"golang-boilerplate/internal/services"
	"golang-boilerplate/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	BaseHandler
	userService services.UserService
	cfg         *config.Config
	validator   *validator.Validate
	restClient  httpclient.RestClient
}

// NewUserHandler creates a new user handler
func ProvideUserHandler(
	userService services.UserService,
	cfg *config.Config,
	validator *validator.Validate,
	restClient httpclient.RestClient,
) *UserHandler {
	return &UserHandler{
		BaseHandler: *NewBaseHandler(),
		userService: userService,
		cfg:         cfg,
		validator:   validator,
		restClient:  restClient,
	}
}

// CreateUser godoc
// @Summary Create user
// @Description Create user
// @Tags User
// @Accept json
// @Produce json
// @Param user body dtos.CreateUserRequest true "User"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.UserResponse}
// @Router /users [post]
// @Security BearerAuth
func (h *UserHandler) CreateUser(c *echo.Context) error {
	_, ok := c.Get(h.cfg.KeycloakKeyClaim).(*auth.TokenClaims)
	if !ok {
		return h.UnauthorizedErrorResponse(c, "User not authenticated")
	}

	var requestDto dtos.CreateUserRequest
	if err := c.Bind(&requestDto); err != nil {
		return h.HandleError(c, errors.ValidationError("Invalid request body", err))
	}

	if err := h.validator.Struct(requestDto); err != nil {
		fieldErrors := errors.ParseValidationErrors(err)
		if len(fieldErrors) > 0 {
			return h.HandleError(c, errors.ValidationErrorWithDetails("Validation failed", err, fieldErrors))
		}
		return h.HandleError(c, errors.ValidationError("Validation failed", err))
	}

	user, err := h.userService.Create(c.Request().Context(), &requestDto)
	if err != nil {
		return h.HandleError(c, err)
	}

	// Return success response with user data
	return h.SuccessResponse(c, "User created successfully", dtos.NewUserResponse(user), nil)
}

// GetOneByID godoc
// @Summary Get user by ID
// @Description Get user by ID
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.UserResponse}
// @Router /users/{id} [get]
// @Security BearerAuth
func (h *UserHandler) GetOneByID(c *echo.Context) error {
	_, ok := c.Get(h.cfg.KeycloakKeyClaim).(*auth.TokenClaims)
	if !ok {
		return h.UnauthorizedErrorResponse(c, "User not authenticated")
	}

	userID := c.Param("id")
	user, err := h.userService.GetOneByID(c.Request().Context(), userID)
	if err != nil {
		return h.HandleError(c, err)
	}

	return h.SuccessResponse(c, "User retrieved successfully", dtos.NewUserResponse(user), nil)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dtos.UpdateUserRequest true "User"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.UserResponse}
// @Router /users/{id} [put]
// @Security BearerAuth
func (h *UserHandler) UpdateUser(c *echo.Context) error {
	_, ok := c.Get(h.cfg.KeycloakKeyClaim).(*auth.TokenClaims)
	if !ok {
		return h.UnauthorizedErrorResponse(c, "User not authenticated")
	}

	userID := c.Param("id")
	var requestDto dtos.UpdateUserRequest
	if err := c.Bind(&requestDto); err != nil {
		return h.HandleError(c, errors.ValidationError("Invalid request body", err))
	}

	if err := h.validator.Struct(requestDto); err != nil {
		fieldErrors := errors.ParseValidationErrors(err)
		if len(fieldErrors) > 0 {
			return h.HandleError(c, errors.ValidationErrorWithDetails("Validation failed", err, fieldErrors))
		}
		return h.HandleError(c, errors.ValidationError("Validation failed", err))
	}

	user, err := h.userService.Update(c.Request().Context(), userID, &requestDto)
	if err != nil {
		return h.HandleError(c, err)
	}

	return h.SuccessResponse(c, "User updated successfully", dtos.NewUserResponse(user), nil)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete user
// @Tags User
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.UserResponse}
// @Router /users/{id} [delete]
// @Security BearerAuth
func (h *UserHandler) DeleteUser(c *echo.Context) error {
	_, ok := c.Get(h.cfg.KeycloakKeyClaim).(*auth.TokenClaims)
	if !ok {
		return h.UnauthorizedErrorResponse(c, "User not authenticated")
	}

	userID := c.Param("id")
	err := h.userService.Delete(c.Request().Context(), userID)
	if err != nil {
		return h.HandleError(c, err)
	}

	return h.SuccessResponse(c, "User deleted successfully", nil, nil)
}

// GetUsers godoc
// @Summary Get users
// @Description Get users
// @Tags User
// @Accept json
// @Produce json
// @Param page query int false "Page" default(1) example("1")
// @Param page_size query int false "Page size" default(10) example("10")
// @Param start_date query string false "Start date" example("2025-09-11T02:17:24.290538Z")
// @Param end_date query string false "End date" example("2025-09-11T02:17:24.290538Z")
// @Param q query string false "Query" example("A")
// @Param sort query string false "Sort" example("[-created_at,name]") Enums(created_at,-created_at,name,-name)
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.UserResponse}
// @Router /users [get]
// @Security BearerAuth
func (h *UserHandler) GetUsers(c *echo.Context) error {
	_, ok := c.Get(h.cfg.KeycloakKeyClaim).(*auth.TokenClaims)
	if !ok {
		return h.UnauthorizedErrorResponse(c, "User not authenticated")
	}

	// Parse pagination parameters
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil || page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.QueryParam("page_size"))
	if err != nil || pageSize < 0 {
		pageSize = 10
	}

	sort := c.QueryParams()["sort"]
	query := c.QueryParam("q")

	// Validate sort fields against allowed set
	allowedSort := map[string]struct{}{
		"created_at": {},
		"name":       {},
	}
	validSort, invalidSort := utils.NormalizeAndValidateSort(sort, allowedSort)
	if len(invalidSort) > 0 {
		invalids := map[string]string{
			"sort": "invalid sort field(s): " + strings.Join(invalidSort, ", "),
		}
		return h.HandleError(c, errors.ValidationErrorWithDetails("Validation failed", nil, invalids))
	}

	// Parse and validate dates
	dateRange, err := utils.ParseDateRange(c.QueryParam("start_date"), c.QueryParam("end_date"))
	if err != nil {
		return h.HandleError(c, errors.ValidationError("Invalid date range", err))
	}

	// Create request DTO
	pr := &dtos.UserPageableRequest{
		PageableRequest: dtos.PageableRequest{
			Page:     page,
			PageSize: pageSize,
		},
		StartDate: dateRange.StartDate,
		EndDate:   dateRange.EndDate,
		Q:         query,
		Sort:      validSort,
	}

	users, err := h.userService.List(c.Request().Context(), pr)
	if err != nil {
		return h.HandleError(c, err)
	}

	// Transform to response DTOs
	responseDto := make([]dtos.UserResponse, len(users.Data))
	for i, user := range users.Data {
		responseDto[i] = *dtos.NewUserResponse(&user)
	}

	return h.SuccessResponse(c, "Users retrieved successfully", responseDto, users.Pageable)
}

// TestRestClient godoc
// @Summary Test rest client
// @Description Test rest client
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} object{meta=dtos.Meta,data=map[string]interface{}}
// @Router /users/test-rest-client [get]
// @Security BearerAuth
func (h *UserHandler) TestRestClient(c *echo.Context) error {
	_, ok := c.Get(h.cfg.KeycloakKeyClaim).(*auth.TokenClaims)
	if !ok {
		return h.UnauthorizedErrorResponse(c, "User not authenticated")
	}

	// Create a result variable to hold the response (JSON array)
	var result []map[string]interface{}
	_, err := h.restClient.Get("https://jsonplaceholder.typicode.com/posts", &result, nil, "")
	if err != nil {
		return h.HandleError(c, err)
	}

	return h.SuccessResponse(c, "User profile retrieved successfully", result, nil)
}
