package catalog

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/midocss/website/internal/domain"
	"github.com/midocss/website/pkg/apperr"
)

type Repository interface {
	ListProjectTypes(ctx context.Context, query ListQuery) ([]domain.ProjectType, int64, error)
	FindProjectType(ctx context.Context, id uuid.UUID) (*domain.ProjectType, error)
	FindProjectTypeBySlug(ctx context.Context, slug string) (*domain.ProjectType, error)
	CreateProjectType(ctx context.Context, projectType *domain.ProjectType) error
	UpdateProjectType(ctx context.Context, id uuid.UUID, fields map[string]any) error
	DeleteProjectType(ctx context.Context, id uuid.UUID) error

	ListPortfolio(ctx context.Context, query ListQuery) ([]domain.PortfolioProject, int64, error)
	FindPortfolioProject(ctx context.Context, id uuid.UUID) (*domain.PortfolioProject, error)
	FindPortfolioProjectBySlug(ctx context.Context, slug string, publishedOnly bool) (*domain.PortfolioProject, error)
	CreatePortfolioProject(ctx context.Context, project *domain.PortfolioProject, images []domain.PortfolioProjectImage) error
	UpdatePortfolioProject(ctx context.Context, id uuid.UUID, fields map[string]any, images *[]domain.PortfolioProjectImage) error
	DeletePortfolioProject(ctx context.Context, id uuid.UUID) error

	ListPackages(ctx context.Context, query ListQuery) ([]domain.Package, int64, error)
	FindPackage(ctx context.Context, id uuid.UUID) (*domain.Package, error)
	FindPackageBySlug(ctx context.Context, slug string, activeOnly bool) (*domain.Package, error)
	CreatePackage(ctx context.Context, pkg *domain.Package, features []domain.PackageFeature) error
	UpdatePackage(ctx context.Context, id uuid.UUID, fields map[string]any, features *[]domain.PackageFeature) error
	DeletePackage(ctx context.Context, id uuid.UUID) error

	CurrencyExists(ctx context.Context, code string) (bool, error)
	// SlugExists checks the given table, ignoring excludeID so an update can
	// keep its own slug.
	SlugExists(ctx context.Context, model any, slug string, excludeID uuid.UUID) (bool, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) ListProjectTypes(ctx context.Context, query ListQuery) ([]domain.ProjectType, int64, error) {
	page, perPage := query.Normalized()

	tx := r.db.WithContext(ctx).Model(&domain.ProjectType{})
	if query.ActiveOnly() {
		tx = tx.Where("is_active")
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		pattern := "%" + strings.ToLower(search) + "%"
		tx = tx.Where("lower(name_ar) LIKE ? OR lower(name_en) LIKE ? OR slug LIKE ?", pattern, pattern, pattern)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, apperr.Internal(err)
	}

	var items []domain.ProjectType
	err := tx.Order("sort_order, name_en").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&items).Error
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}
	return items, total, nil
}

func (r *gormRepository) FindProjectType(ctx context.Context, id uuid.UUID) (*domain.ProjectType, error) {
	var projectType domain.ProjectType
	err := r.db.WithContext(ctx).First(&projectType, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("project type not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &projectType, nil
}

func (r *gormRepository) FindProjectTypeBySlug(ctx context.Context, slug string) (*domain.ProjectType, error) {
	var projectType domain.ProjectType
	err := r.db.WithContext(ctx).First(&projectType, "slug = ?", slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("project type not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &projectType, nil
}

func (r *gormRepository) CreateProjectType(ctx context.Context, projectType *domain.ProjectType) error {
	if projectType.ID == uuid.Nil {
		projectType.ID = uuid.New()
	}
	if err := r.db.WithContext(ctx).Create(projectType).Error; err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *gormRepository) UpdateProjectType(ctx context.Context, id uuid.UUID, fields map[string]any) error {
	return r.update(ctx, &domain.ProjectType{}, id, fields, "project type not found")
}

func (r *gormRepository) DeleteProjectType(ctx context.Context, id uuid.UUID) error {
	return r.softDelete(ctx, &domain.ProjectType{}, id, "project type not found")
}

func (r *gormRepository) ListPortfolio(ctx context.Context, query ListQuery) ([]domain.PortfolioProject, int64, error) {
	page, perPage := query.Normalized()

	tx := r.db.WithContext(ctx).Model(&domain.PortfolioProject{})
	if query.ActiveOnly() {
		tx = tx.Where("portfolio_projects.is_published")
	}
	if query.ProjectTypeSlug != "" {
		tx = tx.Joins("JOIN project_types ON project_types.id = portfolio_projects.project_type_id").
			Where("project_types.slug = ?", query.ProjectTypeSlug)
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		pattern := "%" + strings.ToLower(search) + "%"
		tx = tx.Where("lower(portfolio_projects.title_ar) LIKE ? OR lower(portfolio_projects.title_en) LIKE ?", pattern, pattern)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, apperr.Internal(err)
	}

	var items []domain.PortfolioProject
	err := tx.Preload("ProjectType").
		Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order") }).
		Order("portfolio_projects.sort_order, portfolio_projects.completed_at DESC NULLS LAST").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&items).Error
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}
	return items, total, nil
}

func (r *gormRepository) FindPortfolioProject(ctx context.Context, id uuid.UUID) (*domain.PortfolioProject, error) {
	return r.findPortfolio(ctx, r.db.WithContext(ctx).Where("id = ?", id))
}

func (r *gormRepository) FindPortfolioProjectBySlug(ctx context.Context, slug string, publishedOnly bool) (*domain.PortfolioProject, error) {
	tx := r.db.WithContext(ctx).Where("slug = ?", slug)
	if publishedOnly {
		tx = tx.Where("is_published")
	}
	return r.findPortfolio(ctx, tx)
}

func (r *gormRepository) findPortfolio(ctx context.Context, tx *gorm.DB) (*domain.PortfolioProject, error) {
	var project domain.PortfolioProject
	err := tx.Preload("ProjectType").
		Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order") }).
		First(&project).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("portfolio project not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &project, nil
}

func (r *gormRepository) CreatePortfolioProject(ctx context.Context, project *domain.PortfolioProject, images []domain.PortfolioProjectImage) error {
	if project.ID == uuid.Nil {
		project.ID = uuid.New()
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Images", "ProjectType").Create(project).Error; err != nil {
			return err
		}
		return createImages(tx, project.ID, images)
	})
	if err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *gormRepository) UpdatePortfolioProject(ctx context.Context, id uuid.UUID, fields map[string]any, images *[]domain.PortfolioProjectImage) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(fields) > 0 {
			result := tx.Model(&domain.PortfolioProject{}).Where("id = ?", id).Updates(fields)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return apperr.NotFound("portfolio project not found")
			}
		}
		if images == nil {
			return nil
		}
		if err := tx.Where("portfolio_project_id = ?", id).Delete(&domain.PortfolioProjectImage{}).Error; err != nil {
			return err
		}
		return createImages(tx, id, *images)
	})
	if err != nil {
		return apperr.From(err)
	}
	return nil
}

func createImages(tx *gorm.DB, projectID uuid.UUID, images []domain.PortfolioProjectImage) error {
	if len(images) == 0 {
		return nil
	}
	for i := range images {
		images[i].ID = uuid.New()
		images[i].PortfolioProjectID = projectID
	}
	return tx.Create(&images).Error
}

func (r *gormRepository) DeletePortfolioProject(ctx context.Context, id uuid.UUID) error {
	return r.softDelete(ctx, &domain.PortfolioProject{}, id, "portfolio project not found")
}

func (r *gormRepository) ListPackages(ctx context.Context, query ListQuery) ([]domain.Package, int64, error) {
	page, perPage := query.Normalized()

	tx := r.db.WithContext(ctx).Model(&domain.Package{})
	if query.ActiveOnly() {
		tx = tx.Where("packages.is_active")
	}
	if query.Featured != nil {
		tx = tx.Where("packages.is_featured = ?", *query.Featured)
	}
	if query.ProjectTypeSlug != "" {
		tx = tx.Joins("JOIN project_types ON project_types.id = packages.project_type_id").
			Where("project_types.slug = ?", query.ProjectTypeSlug)
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		pattern := "%" + strings.ToLower(search) + "%"
		tx = tx.Where("lower(packages.name_ar) LIKE ? OR lower(packages.name_en) LIKE ?", pattern, pattern)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, apperr.Internal(err)
	}

	var items []domain.Package
	err := tx.Preload("ProjectType").
		Preload("Features", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order") }).
		Order("packages.sort_order, packages.price_amount").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&items).Error
	if err != nil {
		return nil, 0, apperr.Internal(err)
	}
	return items, total, nil
}

func (r *gormRepository) FindPackage(ctx context.Context, id uuid.UUID) (*domain.Package, error) {
	return r.findPackage(ctx, r.db.WithContext(ctx).Where("id = ?", id))
}

func (r *gormRepository) FindPackageBySlug(ctx context.Context, slug string, activeOnly bool) (*domain.Package, error) {
	tx := r.db.WithContext(ctx).Where("slug = ?", slug)
	if activeOnly {
		tx = tx.Where("is_active")
	}
	return r.findPackage(ctx, tx)
}

func (r *gormRepository) findPackage(ctx context.Context, tx *gorm.DB) (*domain.Package, error) {
	var pkg domain.Package
	err := tx.Preload("ProjectType").
		Preload("Features", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order") }).
		First(&pkg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.NotFound("package not found")
	}
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &pkg, nil
}

func (r *gormRepository) CreatePackage(ctx context.Context, pkg *domain.Package, features []domain.PackageFeature) error {
	if pkg.ID == uuid.Nil {
		pkg.ID = uuid.New()
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Features", "ProjectType").Create(pkg).Error; err != nil {
			return err
		}
		return createFeatures(tx, pkg.ID, features)
	})
	if err != nil {
		return apperr.Internal(err)
	}
	return nil
}

func (r *gormRepository) UpdatePackage(ctx context.Context, id uuid.UUID, fields map[string]any, features *[]domain.PackageFeature) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(fields) > 0 {
			result := tx.Model(&domain.Package{}).Where("id = ?", id).Updates(fields)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return apperr.NotFound("package not found")
			}
		}
		if features == nil {
			return nil
		}
		if err := tx.Where("package_id = ?", id).Delete(&domain.PackageFeature{}).Error; err != nil {
			return err
		}
		return createFeatures(tx, id, *features)
	})
	if err != nil {
		return apperr.From(err)
	}
	return nil
}

func createFeatures(tx *gorm.DB, packageID uuid.UUID, features []domain.PackageFeature) error {
	if len(features) == 0 {
		return nil
	}
	for i := range features {
		features[i].ID = uuid.New()
		features[i].PackageID = packageID
	}
	return tx.Create(&features).Error
}

func (r *gormRepository) DeletePackage(ctx context.Context, id uuid.UUID) error {
	return r.softDelete(ctx, &domain.Package{}, id, "package not found")
}

func (r *gormRepository) CurrencyExists(ctx context.Context, code string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Currency{}).
		Where("code = ? AND is_active", strings.ToUpper(code)).
		Count(&count).Error
	if err != nil {
		return false, apperr.Internal(err)
	}
	return count > 0, nil
}

func (r *gormRepository) SlugExists(ctx context.Context, model any, slug string, excludeID uuid.UUID) (bool, error) {
	tx := r.db.WithContext(ctx).Model(model).Where("slug = ?", slug)
	if excludeID != uuid.Nil {
		tx = tx.Where("id <> ?", excludeID)
	}

	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return false, apperr.Internal(err)
	}
	return count > 0, nil
}

func (r *gormRepository) update(ctx context.Context, model any, id uuid.UUID, fields map[string]any, notFound string) error {
	if len(fields) == 0 {
		return nil
	}
	result := r.db.WithContext(ctx).Model(model).Where("id = ?", id).Updates(fields)
	if result.Error != nil {
		return apperr.Internal(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound(notFound)
	}
	return nil
}

func (r *gormRepository) softDelete(ctx context.Context, model any, id uuid.UUID, notFound string) error {
	result := r.db.WithContext(ctx).Delete(model, "id = ?", id)
	if result.Error != nil {
		return apperr.Internal(result.Error)
	}
	if result.RowsAffected == 0 {
		return apperr.NotFound(notFound)
	}
	return nil
}
