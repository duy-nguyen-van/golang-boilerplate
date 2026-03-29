package repositories

import (
	"golang-boilerplate/internal/db"
	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/models"
	"strings"

	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *models.User) (*models.User, error)
	GetOneByID(id string, preloads ...string) (*models.User, error)
	Update(user *models.User) error
	Delete(user *models.User) error
	Get(pr *dtos.UserPageableRequest, preloads ...string) (*dtos.DataResponse[models.User], error)
}

// userRepository implements UserRepository
type userRepository struct {
	abstractRepository[models.User]
}

// ProvideUserRepository creates a new user repository
func ProvideUserRepository(db *db.PostgresDB) UserRepository {
	return &userRepository{
		abstractRepository: abstractRepository[models.User]{db: db},
	}
}

func (r *userRepository) Create(user *models.User) (*models.User, error) {
	err := r.Save(user)
	if err != nil {
		return nil, errors.DatabaseError("Failed to create user", err).
			WithOperation("create_user").
			WithResource("user")
	}

	return user, nil
}

func (r *userRepository) GetOneByID(id string, preloads ...string) (*models.User, error) {
	query := r.db.DB
	user := &models.User{}

	if len(preloads) > 0 {
		for _, preload := range preloads {
			query = query.Preload(preload)
		}
	}
	query = query.Where("id = ?", id)

	err := query.First(user).Error
	if err != nil {
		return nil, errors.DatabaseError("Failed to get user by ID", err).
			WithOperation("get_user_by_id").
			WithResource("user").
			WithContext("user_id", id)
	}

	return user, nil
}

func (r *userRepository) Update(user *models.User) error {
	// Use a transaction to ensure atomicity
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// First update the user fields
		result := tx.Updates(user)
		if result.Error != nil {
			return result.Error
		}

		// Save the companies we want to associate before clearing
		companiesToAssociate := user.Companies

		// Clear existing associations first
		if err := tx.Model(user).Association("Companies").Clear(); err != nil {
			return err
		}

		// Then add the new associations
		if len(companiesToAssociate) > 0 {
			if err := tx.Model(user).Association("Companies").Append(companiesToAssociate); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return errors.DatabaseError("Failed to update user", err).
			WithOperation("update_user").
			WithResource("user").
			WithContext("user_id", user.ID)
	}

	return nil
}

func (r *userRepository) Delete(user *models.User) error {
	result := r.db.Delete(user)
	if result.Error != nil {
		return errors.DatabaseError("Failed to delete user", result.Error).
			WithOperation("delete_user").
			WithResource("user").
			WithContext("user_id", user.ID)
	}

	return nil
}

func (r *userRepository) Get(pr *dtos.UserPageableRequest, preloads ...string) (*dtos.DataResponse[models.User], error) {
	query := r.db.DB

	// Apply preloading if set
	if len(preloads) > 0 {
		for _, preload := range preloads {
			query = query.Preload(preload)
		}
	}

	if pr.Q != "" {
		query = query.Where("users.first_name LIKE ?", "%"+pr.Q+"%")
		query = query.Or("users.last_name LIKE ?", "%"+pr.Q+"%")
		query = query.Or("users.email LIKE ?", "%"+pr.Q+"%")
	}

	if pr.StartDate != nil {
		query = query.Where("users.created_at >= ?", pr.StartDate)
	}

	if pr.EndDate != nil {
		query = query.Where("users.created_at <= ?", pr.EndDate)
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
				query = query.Order("users." + sortField + " desc")
			} else {
				query = query.Order("users." + sortField + " asc")
			}

		}
	} else {
		query = query.Order("users.created_at desc")
	}

	result, err := r.find(query, &pr.PageableRequest)
	if err != nil {
		return nil, errors.DatabaseError("Failed to get users", err).
			WithOperation("get_users").
			WithResource("users").
			WithContext("pageable_request", pr)
	}

	return result, nil
}
