package handlers

import (
	"strconv"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/integration/auth"
	"golang-boilerplate/internal/services"
	"golang-boilerplate/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

// CompanyHandler handles company-related HTTP requests
type CompanyHandler struct {
	BaseHandler
	companyService services.CompanyService
	cfg            *config.Config
	validator      *validator.Validate
}

// ProvideCompanyHandler creates a new company handler
func ProvideCompanyHandler(
	companyService services.CompanyService,
	cfg *config.Config,
	validator *validator.Validate,
) *CompanyHandler {
	return &CompanyHandler{
		BaseHandler:    *NewBaseHandler(),
		companyService: companyService,
		cfg:            cfg,
		validator:      validator,
	}
}

// CreateCompany godoc
// @Summary Create company
// @Description Create company
// @Tags Company
// @Accept json
// @Produce json
// @Param company body dtos.CreateCompanyRequest true "Company"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CompanyResponse}
// @Router /companies [post]
// @Security BearerAuth
func (h *CompanyHandler) CreateCompany(c *echo.Context) error {
	_, ok := c.Get(h.cfg.KeycloakKeyClaim).(*auth.TokenClaims)
	if !ok {
		return h.UnauthorizedErrorResponse(c, "User not authenticated")
	}

	var requestDto dtos.CreateCompanyRequest
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

	company, err := h.companyService.Create(c.Request().Context(), &requestDto)
	if err != nil {
		return h.HandleError(c, err)
	}

	// Return success response with company data
	return h.SuccessResponse(c, "Company created successfully", dtos.NewCompanyResponse(company), nil)
}

// GetOneByID godoc
// @Summary Get company by ID
// @Description Get company by ID
// @Tags Company
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CompanyResponse}
// @Router /companies/{id} [get]
// @Security BearerAuth
func (h *CompanyHandler) GetOneByID(c *echo.Context) error {
	_, ok := c.Get(h.cfg.KeycloakKeyClaim).(*auth.TokenClaims)
	if !ok {
		return h.UnauthorizedErrorResponse(c, "User not authenticated")
	}

	companyID := c.Param("id")
	company, err := h.companyService.GetOneByID(c.Request().Context(), companyID)
	if err != nil {
		return h.HandleError(c, err)
	}

	return h.SuccessResponse(c, "Company retrieved successfully", dtos.NewCompanyResponse(company), nil)
}

// UpdateCompany godoc
// @Summary Update company
// @Description Update company
// @Tags Company
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Param company body dtos.UpdateCompanyRequest true "Company"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CompanyResponse}
// @Router /companies/{id} [put]
// @Security BearerAuth
func (h *CompanyHandler) UpdateCompany(c *echo.Context) error {
	_, ok := c.Get(h.cfg.KeycloakKeyClaim).(*auth.TokenClaims)
	if !ok {
		return h.UnauthorizedErrorResponse(c, "User not authenticated")
	}

	companyID := c.Param("id")
	var requestDto dtos.UpdateCompanyRequest
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

	company, err := h.companyService.Update(c.Request().Context(), companyID, &requestDto)
	if err != nil {
		return h.HandleError(c, err)
	}

	return h.SuccessResponse(c, "Company updated successfully", dtos.NewCompanyResponse(company), nil)
}

// DeleteCompany godoc
// @Summary Delete company
// @Description Delete company
// @Tags Company
// @Accept json
// @Produce json
// @Param id path string true "Company ID"
// @Success 200 {object} object{meta=dtos.Meta,data=dtos.CompanyResponse}
// @Router /companies/{id} [delete]
// @Security BearerAuth
func (h *CompanyHandler) DeleteCompany(c *echo.Context) error {
	_, ok := c.Get(h.cfg.KeycloakKeyClaim).(*auth.TokenClaims)
	if !ok {
		return h.UnauthorizedErrorResponse(c, "User not authenticated")
	}

	companyID := c.Param("id")
	err := h.companyService.Delete(c.Request().Context(), companyID)
	if err != nil {
		return h.HandleError(c, err)
	}

	return h.SuccessResponse(c, "Company deleted successfully", nil, nil)
}

// GetCompanies godoc
// @Summary Get companies
// @Description Get companies
// @Tags Company
// @Accept json
// @Produce json
// @Param page query int false "Page" default(1) example("1")
// @Param page_size query int false "Page size" default(10) example("10")
// @Param start_date query string false "Start date" example("2025-09-11T02:17:24.290538Z")
// @Param end_date query string false "End date" example("2025-09-11T02:17:24.290538Z")
// @Param q query string false "Query" example("A")
// @Param sort query string false "Sort" example("[-created_at,name]") Enums(created_at,-created_at,name,-name)
// @Success 200 {object} object{meta=dtos.Meta,data=[]dtos.CompanyResponse}
// @Router /companies [get]
// @Security BearerAuth
func (h *CompanyHandler) GetCompanies(c *echo.Context) error {
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

	// Parse and validate dates
	dateRange, err := utils.ParseDateRange(c.QueryParam("start_date"), c.QueryParam("end_date"))
	if err != nil {
		return h.HandleError(c, errors.ValidationError("Invalid date range", err))
	}

	// Create request DTO
	pr := &dtos.CompanyPageableRequest{
		PageableRequest: dtos.PageableRequest{
			Page:     page,
			PageSize: pageSize,
		},
		StartDate: dateRange.StartDate,
		EndDate:   dateRange.EndDate,
		Q:         query,
		Sort:      sort,
	}

	companies, err := h.companyService.List(c.Request().Context(), pr)
	if err != nil {
		return h.HandleError(c, err)
	}

	// Transform to response DTOs
	responseDto := make([]dtos.CompanyResponse, len(companies.Data))
	for i, company := range companies.Data {
		responseDto[i] = *dtos.NewCompanyResponse(&company)
	}

	return h.SuccessResponse(c, "Companies retrieved successfully", responseDto, companies.Pageable)
}
