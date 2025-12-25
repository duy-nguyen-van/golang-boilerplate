package services

import (
	"context"

	"golang-boilerplate/internal/cache"
	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/models"
	"golang-boilerplate/internal/repositories"

	"golang-boilerplate/internal/logger"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

type CompanyService interface {
	Create(ctx context.Context, req *dtos.CreateCompanyRequest) (*models.Company, error)
	GetOneByID(ctx context.Context, companyID string) (*models.Company, error)
	Update(ctx context.Context, companyID string, req *dtos.UpdateCompanyRequest) (*models.Company, error)
	Delete(ctx context.Context, companyID string) error
	List(ctx context.Context, pageableRequest *dtos.CompanyPageableRequest) (*dtos.DataResponse[models.Company], error)
}

// CompanyService handles company business logic
type companyService struct {
	companyRepo repositories.CompanyRepository
	cache       cache.Cache
}

// ProvideCompanyService creates a new company service
func ProvideCompanyService(
	companyRepo repositories.CompanyRepository,
	cache cache.Cache,
) CompanyService {
	return &companyService{
		companyRepo: companyRepo,
		cache:       cache,
	}
}

func (s *companyService) Create(ctx context.Context, req *dtos.CreateCompanyRequest) (*models.Company, error) {
	company := &models.Company{
		Name:       req.Name,
		KeycloakID: req.KeycloakID,
	}

	company, err := s.companyRepo.Create(company)
	if err != nil {
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("service", "company_service")
				scope.SetTag("operation", "create_company")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("body_request", req)
				hub.CaptureException(err)
			})
		}

		logger.Log.Error("Failed to create company",
			zap.Any("body_request", req),
			zap.Error(err),
		)

		return nil, errors.DatabaseError("Failed to create company", err).
			WithOperation("create_company").
			WithResource("company").
			WithContext("request", req)
	}

	return company, nil
}

func (s *companyService) GetOneByID(ctx context.Context, companyID string) (*models.Company, error) {
	company, err := s.companyRepo.GetOneByID(companyID)
	if err != nil {
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("service", "company_service")
				scope.SetTag("operation", "get_company")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("company_id", companyID)
				hub.CaptureException(err)
			})
		}

		logger.Log.Error("Failed to get company",
			zap.String("company_id", companyID),
			zap.Error(err),
		)

		return nil, errors.NotFoundError("Company", err).
			WithOperation("get_company").
			WithResource("company").
			WithContext("company_id", companyID)
	}

	return company, nil
}

func (s *companyService) Update(ctx context.Context, companyID string, req *dtos.UpdateCompanyRequest) (*models.Company, error) {
	company, err := s.companyRepo.GetOneByID(companyID)
	if err != nil {
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("service", "company_service")
				scope.SetTag("operation", "update_company")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("company_id", companyID)
				scope.SetExtra("body_request", req)
				hub.CaptureException(err)
			})
		}

		logger.Log.Error("Failed to get company for update",
			zap.String("company_id", companyID),
			zap.Any("body_request", req),
			zap.Error(err),
		)

		return nil, errors.NotFoundError("Company", err).
			WithOperation("update_company").
			WithResource("company").
			WithContext("company_id", companyID)
	}

	// Update company fields if provided in request
	if req.Name != "" {
		company.Name = req.Name
	}
	if req.KeycloakID != "" {
		company.KeycloakID = req.KeycloakID
	}

	// Save the updated company
	err = s.companyRepo.Update(company)
	if err != nil {
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("service", "company_service")
				scope.SetTag("operation", "update_company")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("company_id", companyID)
				scope.SetExtra("body_request", req)
				hub.CaptureException(err)
			})
		}

		logger.Log.Error("Failed to update company",
			zap.String("company_id", companyID),
			zap.Any("body_request", req),
			zap.Error(err),
		)

		return nil, errors.DatabaseError("Failed to update company", err).
			WithOperation("update_company").
			WithResource("company").
			WithContext("company_id", companyID)
	}

	return company, nil
}

func (s *companyService) Delete(ctx context.Context, companyID string) error {
	company, err := s.companyRepo.GetOneByID(companyID)
	if err != nil {
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("service", "company_service")
				scope.SetTag("operation", "delete_company")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("company_id", companyID)
				hub.CaptureException(err)
			})
		}

		logger.Log.Error("Failed to get company for delete",
			zap.String("company_id", companyID),
			zap.Error(err),
		)

		return errors.NotFoundError("Company", err).
			WithOperation("delete_company").
			WithResource("company").
			WithContext("company_id", companyID)
	}

	err = s.companyRepo.Delete(company)
	if err != nil {
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("service", "company_service")
				scope.SetTag("operation", "delete_company")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("company_id", companyID)
				hub.CaptureException(err)
			})
		}

		logger.Log.Error("Failed to delete company",
			zap.String("company_id", companyID),
			zap.Error(err),
		)

		return errors.DatabaseError("Failed to delete company", err).
			WithOperation("delete_company").
			WithResource("company").
			WithContext("company_id", companyID)
	}

	return nil
}

func (s *companyService) List(ctx context.Context, pageableRequest *dtos.CompanyPageableRequest) (*dtos.DataResponse[models.Company], error) {
	companies, err := s.companyRepo.Get(pageableRequest)
	if err != nil {
		if hub := sentry.GetHubFromContext(ctx); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.SetTag("service", "company_service")
				scope.SetTag("operation", "get_companies")
				scope.SetExtra("error_details", err.Error())
				scope.SetExtra("pageable_request", pageableRequest)
				hub.CaptureException(err)
			})
		}

		logger.Log.Error("Failed to get companies",
			zap.Any("pageable_request", pageableRequest),
			zap.Error(err),
		)

		return nil, errors.DatabaseError("Failed to get companies", err).
			WithOperation("get_companies").
			WithResource("companies").
			WithContext("pageable_request", pageableRequest)
	}

	return companies, nil
}
