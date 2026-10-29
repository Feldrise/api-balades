package dbmodel

import (
	"context"
	"fmt"
	"time"

	"feldrise.com/balade/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type User struct {
	gorm.Model

	Email string `gorm:"not null;unique"`

	AuthenticationCode     *string
	AuthenticationExpireAt *time.Time

	Roles               []Role                   `gorm:"many2many:user_roles;"`
	PermissionOverrides []UserPermissionOverride `gorm:"foreignKey:UserID"`
	UserProfile         UserProfile              `gorm:"foreignKey:UserID"`
	Guide               *Guide                   `gorm:"foreignKey:UserID"`
}

func (u *User) ToModel() *model.User {
	var permissionsSet = make(map[string]bool)
	var roles []string
	var profile *model.UserProfile

	if len(u.Roles) > 0 {
		for _, role := range u.Roles {
			if len(role.Permissions) > 0 {
				for _, permission := range role.Permissions {
					permissionsSet[permission.Name] = true
				}
			}

			roles = append(roles, role.Name)
		}
	}

	if len(u.PermissionOverrides) > 0 {
		for _, override := range u.PermissionOverrides {
			if override.IsGranted {
				permissionsSet[override.Permission.Name] = true
			} else {
				permissionsSet[override.Permission.Name] = false
			}
		}
	}

	var permissions []string
	for permission, granted := range permissionsSet {
		if granted {
			permissions = append(permissions, permission)
		}
	}

	if u.UserProfile.ID != 0 {
		profile = u.UserProfile.ToModel()
	}

	return &model.User{
		ID:          u.ID,
		CreatedAt:   u.CreatedAt,
		Email:       u.Email,
		Permissions: permissions,
		Roles:       roles,
		Profile:     profile,
	}
}

func (u *User) HasPermission(permission string) bool {
	for _, override := range u.PermissionOverrides {
		if override.Permission.Name == permission {
			return override.IsGranted
		}
	}

	for _, role := range u.Roles {
		for _, p := range role.Permissions {
			if p.Name == permission {
				return true
			}
		}
	}

	return false
}

type UserFieldsToInclude struct {
	UserProfile       bool
	Roles             bool
	Roles_Permissions bool
}

type UserFilter struct {
	Search *string
	Limit  int
}

type UserRepository interface {
	FindByID(id uint, fieldsToInclude *UserFieldsToInclude) (*User, error)
	// FindByEmail finds the account of email, whatever its case.
	FindByEmail(email string, fieldsToInclude *UserFieldsToInclude) (*User, error)
	Search(filter *UserFilter) ([]User, error)
	Create(user *User) (*User, error)
	// FindOrCreateByEmail returns the account of user.Email, whatever its case,
	// or creates user with its profile. Concurrent calls for the same email end
	// up with the same single account.
	FindOrCreateByEmail(user *User) (*User, error)
	Update(user *User) (*User, error)
	Delete(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(id uint, fieldsToInclude *UserFieldsToInclude) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	tx := r.db.WithContext(ctx).Model(&user)

	if fieldsToInclude != nil {
		if fieldsToInclude.UserProfile {
			tx = tx.Preload("UserProfile")
		}

		if fieldsToInclude.Roles {
			tx = tx.Preload("Roles")
		}

		// if fieldsToInclude.Roles_Permissions {
		// 	tx = tx.Preload("Roles.Permissions").Preload("PermissionOverrides").Preload("PermissionOverrides.Permission")
		// }
	}
	tx = tx.Preload("Roles.Permissions").Preload("PermissionOverrides").Preload("PermissionOverrides.Permission")

	err := tx.Where("id = ?", id).First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Search(filter *UserFilter) ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var users []User
	tx := r.db.WithContext(ctx).Model(&User{}).Preload("UserProfile")

	if filter != nil {
		if filter.Search != nil && *filter.Search != "" {
			search := "%" + *filter.Search + "%"
			tx = tx.Joins("LEFT JOIN user_profiles ON user_profiles.user_id = users.id AND user_profiles.deleted_at IS NULL").
				Where(
					"users.email ILIKE ? OR user_profiles.first_name ILIKE ? OR user_profiles.last_name ILIKE ?",
					search, search, search,
				)
		}

		limit := filter.Limit
		if limit <= 0 {
			limit = 20
		}
		tx = tx.Limit(limit)
	} else {
		tx = tx.Limit(20)
	}

	err := tx.Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) FindByEmail(email string, fieldsToInclude *UserFieldsToInclude) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	tx := r.db.WithContext(ctx).Model(&user)

	if fieldsToInclude != nil {
		if fieldsToInclude.UserProfile {
			tx = tx.Preload("UserProfile")
		}

		if fieldsToInclude.Roles {
			tx = tx.Preload("Roles")
		}

		if fieldsToInclude.Roles_Permissions {
			tx = tx.Preload("Roles.Permissions")
		}
	}

	err := tx.Where("LOWER(email) = LOWER(?)", email).Order("id ASC").First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Create(user *User) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx := r.db.WithContext(ctx)

	err := tx.Create(user).Error

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) FindOrCreateByEmail(user *User) (*User, error) {
	existing, err := r.FindByEmail(user.Email, nil)
	if err != nil || existing != nil {
		return existing, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		profile := user.UserProfile

		// Waits for a concurrent creation of the same email and then does nothing.
		result := tx.Omit(clause.Associations).Clauses(clause.OnConflict{DoNothing: true}).Create(user)
		if result.Error != nil || result.RowsAffected == 0 {
			return result.Error
		}

		profile.UserID = user.ID
		if err := tx.Create(&profile).Error; err != nil {
			return err
		}
		user.UserProfile = profile
		return nil
	})
	if err != nil {
		return nil, err
	}
	if user.ID != 0 {
		return user, nil
	}

	existing, err = r.FindByEmail(user.Email, nil)
	if err == nil && existing == nil {
		// The conflicting account is soft deleted.
		return nil, fmt.Errorf("the account of %s is deleted", user.Email)
	}
	return existing, err
}

func (r *userRepository) Update(user *User) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx := r.db.Session(&gorm.Session{FullSaveAssociations: true}).WithContext(ctx)
	// tx := r.db.WithContext(ctx)

	err := tx.Save(user).Error

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) Delete(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx := r.db.WithContext(ctx)

	err := tx.Select(clause.Associations).Delete(&User{}, id).Error

	if err != nil {
		return err
	}

	return nil
}
