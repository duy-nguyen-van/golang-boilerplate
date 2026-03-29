package repositories

import (
	"golang-boilerplate/internal/db"
	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/models"
	"strings"
)

// CompanyRepository defines the interface for company data operations
type CompanyRepository interface {
	Create(company *models.Company) (*models.Company, error)
	GetOneByID(id string) (*models.Company, error)
	Update(company *models.Company) error
	Delete(company *models.Company) error
	Get(pr *dtos.CompanyPageableRequest, preloads ...string) (*dtos.DataResponse[models.Company], error)
}

// companyRepository implements CompanyRepository
type companyRepository struct {
	abstractRepository[models.Company]
}

// ProvideCompanyRepository creates a new company repository
func ProvideCompanyRepository(db *db.PostgresDB) CompanyRepository {
	return &companyRepository{
		abstractRepository: abstractRepository[models.Company]{db: db},
	}
}

func (r *companyRepository) Create(company *models.Company) (*models.Company, error) {
	err := r.Save(company)
	if err != nil {
		return nil, errors.DatabaseError("Failed to create company", err).
			WithOperation("create_company").
			WithResource("company")
	}

	return company, nil
}

func (r *companyRepository) GetOneByID(id string) (*models.Company, error) {
	company, err := r.FindOneByID(id)
	if err != nil {
		return nil, errors.DatabaseError("Failed to get company by ID", err).
			WithOperation("get_company_by_id").
			WithResource("company").
			WithContext("company_id", id)
	}

	return company, nil
}

func (r *companyRepository) Update(company *models.Company) error {
	result := r.db.Updates(company)
	if result.Error != nil {
		return errors.DatabaseError("Failed to update company", result.Error).
			WithOperation("update_company").
			WithResource("company").
			WithContext("company_id", company.ID)
	}

	return nil
}

func (r *companyRepository) Delete(company *models.Company) error {
	result := r.db.Delete(company)
	if result.Error != nil {
		return errors.DatabaseError("Failed to delete company", result.Error).
			WithOperation("delete_company").
			WithResource("company").
			WithContext("company_id", company.ID)
	}

	return nil
}

func (r *companyRepository) Get(pr *dtos.CompanyPageableRequest, preloads ...string) (*dtos.DataResponse[models.Company], error) {
	query := r.db.DB

	// Apply preloading if set
	if len(preloads) > 0 {
		for _, preload := range preloads {
			query = query.Preload(preload)
		}
	}

	if pr.Q != "" {
		query = query.Where("companies.name LIKE ?", "%"+pr.Q+"%")
	}

	if pr.StartDate != nil {
		query = query.Where("companies.created_at >= ?", pr.StartDate)
	}

	if pr.EndDate != nil {
		query = query.Where("companies.created_at <= ?", pr.EndDate)
	}

	// Apply multiple sort criteria
	if len(pr.Sort) > 0 {
		for _, field := range pr.Sort {
			isDesc := false
			sortField := field

			if strings.HasPrefix(field, "-") {
				isDesc = true
				sortField = strings.TrimPrefix(field, "-")
			}

			if isDesc {
				query = query.Order("companies." + sortField + " desc")
			} else {
				query = query.Order("companies." + sortField + " asc")
			}

		}
	} else {
		query = query.Order("companies.created_at desc")
	}

	result, err := r.find(query, &pr.PageableRequest)
	if err != nil {
		return nil, errors.DatabaseError("Failed to get companies", err).
			WithOperation("get_companies").
			WithResource("companies").
			WithContext("pageable_request", pr)
	}

	return result, nil
}
