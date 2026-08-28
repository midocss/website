package catalog

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/midocss/website/internal/domain"
	"github.com/midocss/website/pkg/apperr"
)

// Service owns the catalog rules shared by the dashboard and the public site:
// slug generation and uniqueness, price/currency validation and the
// published/active visibility of every entity.
type Service interface {
	ListProjectTypes(ctx context.Context, query ListQuery) ([]domain.ProjectType, int64, error)
	GetProjectType(ctx context.Context, id uuid.UUID) (*domain.ProjectType, error)
	GetProjectTypeBySlug(ctx context.Context, slug string) (*domain.ProjectType, error)
	CreateProjectType(ctx context.Context, in CreateProjectTypeInput) (*domain.ProjectType, error)
	UpdateProjectType(ctx context.Context, id uuid.UUID, in UpdateProjectTypeInput) (*domain.ProjectType, error)
	DeleteProjectType(ctx context.Context, id uuid.UUID) error

	ListPortfolio(ctx context.Context, query ListQuery) ([]domain.PortfolioProject, int64, error)
	GetPortfolioProject(ctx context.Context, id uuid.UUID) (*domain.PortfolioProject, error)
	GetPortfolioProjectBySlug(ctx context.Context, slug string, publishedOnly bool) (*domain.PortfolioProject, error)
	CreatePortfolioProject(ctx context.Context, in CreatePortfolioProjectInput) (*domain.PortfolioProject, error)
	UpdatePortfolioProject(ctx context.Context, id uuid.UUID, in UpdatePortfolioProjectInput) (*domain.PortfolioProject, error)
	DeletePortfolioProject(ctx context.Context, id uuid.UUID) error

	ListPackages(ctx context.Context, query ListQuery) ([]domain.Package, int64, error)
	GetPackage(ctx context.Context, id uuid.UUID) (*domain.Package, error)
	GetPackageBySlug(ctx context.Context, slug string, activeOnly bool) (*domain.Package, error)
	CreatePackage(ctx context.Context, in CreatePackageInput) (*domain.Package, error)
	UpdatePackage(ctx context.Context, id uuid.UUID, in UpdatePackageInput) (*domain.Package, error)
	DeletePackage(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListProjectTypes(ctx context.Context, query ListQuery) ([]domain.ProjectType, int64, error) {
	return s.repo.ListProjectTypes(ctx, query)
}

func (s *service) GetProjectType(ctx context.Context, id uuid.UUID) (*domain.ProjectType, error) {
	return s.repo.FindProjectType(ctx, id)
}

func (s *service) GetProjectTypeBySlug(ctx context.Context, slug string) (*domain.ProjectType, error) {
	return s.repo.FindProjectTypeBySlug(ctx, slug)
}

func (s *service) CreateProjectType(ctx context.Context, in CreateProjectTypeInput) (*domain.ProjectType, error) {
	slug, err := s.resolveSlug(ctx, &domain.ProjectType{}, in.Slug, in.NameEn, in.NameAr, uuid.Nil)
	if err != nil {
		return nil, err
	}

	projectType := &domain.ProjectType{
		ID:            uuid.New(),
		Slug:          slug,
		NameAr:        strings.TrimSpace(in.NameAr),
		NameEn:        strings.TrimSpace(in.NameEn),
		DescriptionAr: in.DescriptionAr,
		DescriptionEn: in.DescriptionEn,
		ColorHex:      strings.ToLower(in.ColorHex),
		IconObjectKey: in.IconObjectKey,
		SortOrder:     in.SortOrder,
		IsActive:      boolOr(in.IsActive, true),
	}
	if err := s.repo.CreateProjectType(ctx, projectType); err != nil {
		return nil, err
	}
	return projectType, nil
}

func (s *service) UpdateProjectType(ctx context.Context, id uuid.UUID, in UpdateProjectTypeInput) (*domain.ProjectType, error) {
	fields := map[string]any{}
	if in.Slug != nil {
		slug, err := s.resolveSlug(ctx, &domain.ProjectType{}, *in.Slug, "", "", id)
		if err != nil {
			return nil, err
		}
		fields["slug"] = slug
	}
	setString(fields, "name_ar", in.NameAr)
	setString(fields, "name_en", in.NameEn)
	setPointer(fields, "description_ar", in.DescriptionAr)
	setPointer(fields, "description_en", in.DescriptionEn)
	setPointer(fields, "icon_object_key", in.IconObjectKey)
	if in.ColorHex != nil {
		fields["color_hex"] = strings.ToLower(*in.ColorHex)
	}
	if in.SortOrder != nil {
		fields["sort_order"] = *in.SortOrder
	}
	if in.IsActive != nil {
		fields["is_active"] = *in.IsActive
	}

	if err := s.repo.UpdateProjectType(ctx, id, fields); err != nil {
		return nil, err
	}
	return s.repo.FindProjectType(ctx, id)
}

func (s *service) DeleteProjectType(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteProjectType(ctx, id)
}

func (s *service) ListPortfolio(ctx context.Context, query ListQuery) ([]domain.PortfolioProject, int64, error) {
	return s.repo.ListPortfolio(ctx, query)
}

func (s *service) GetPortfolioProject(ctx context.Context, id uuid.UUID) (*domain.PortfolioProject, error) {
	return s.repo.FindPortfolioProject(ctx, id)
}

func (s *service) GetPortfolioProjectBySlug(ctx context.Context, slug string, publishedOnly bool) (*domain.PortfolioProject, error) {
	return s.repo.FindPortfolioProjectBySlug(ctx, slug, publishedOnly)
}

func (s *service) CreatePortfolioProject(ctx context.Context, in CreatePortfolioProjectInput) (*domain.PortfolioProject, error) {
	slug, err := s.resolveSlug(ctx, &domain.PortfolioProject{}, in.Slug, in.TitleEn, in.TitleAr, uuid.Nil)
	if err != nil {
		return nil, err
	}
	projectTypeID, err := s.resolveProjectType(ctx, in.ProjectTypeID)
	if err != nil {
		return nil, err
	}
	completedAt, err := parseDate(in.CompletedAt)
	if err != nil {
		return nil, err
	}

	project := &domain.PortfolioProject{
		ID:             uuid.New(),
		ProjectTypeID:  projectTypeID,
		Slug:           slug,
		TitleAr:        strings.TrimSpace(in.TitleAr),
		TitleEn:        strings.TrimSpace(in.TitleEn),
		DescriptionAr:  in.DescriptionAr,
		DescriptionEn:  in.DescriptionEn,
		ExternalURL:    in.ExternalURL,
		CoverObjectKey: in.CoverObjectKey,
		CompletedAt:    completedAt,
		SortOrder:      in.SortOrder,
		IsPublished:    boolOr(in.IsPublished, false),
	}
	if err := s.repo.CreatePortfolioProject(ctx, project, toImages(in.Images)); err != nil {
		return nil, err
	}
	return s.repo.FindPortfolioProject(ctx, project.ID)
}

func (s *service) UpdatePortfolioProject(ctx context.Context, id uuid.UUID, in UpdatePortfolioProjectInput) (*domain.PortfolioProject, error) {
	fields := map[string]any{}
	if in.Slug != nil {
		slug, err := s.resolveSlug(ctx, &domain.PortfolioProject{}, *in.Slug, "", "", id)
		if err != nil {
			return nil, err
		}
		fields["slug"] = slug
	}
	if in.ProjectTypeID != nil {
		projectTypeID, err := s.resolveProjectType(ctx, in.ProjectTypeID)
		if err != nil {
			return nil, err
		}
		fields["project_type_id"] = projectTypeID
	}
	if in.CompletedAt != nil {
		completedAt, err := parseDate(in.CompletedAt)
		if err != nil {
			return nil, err
		}
		fields["completed_at"] = completedAt
	}
	setString(fields, "title_ar", in.TitleAr)
	setString(fields, "title_en", in.TitleEn)
	setPointer(fields, "description_ar", in.DescriptionAr)
	setPointer(fields, "description_en", in.DescriptionEn)
	setPointer(fields, "external_url", in.ExternalURL)
	setPointer(fields, "cover_object_key", in.CoverObjectKey)
	if in.SortOrder != nil {
		fields["sort_order"] = *in.SortOrder
	}
	if in.IsPublished != nil {
		fields["is_published"] = *in.IsPublished
	}

	var images *[]domain.PortfolioProjectImage
	if in.Images != nil {
		converted := toImages(*in.Images)
		images = &converted
	}

	if err := s.repo.UpdatePortfolioProject(ctx, id, fields, images); err != nil {
		return nil, err
	}
	return s.repo.FindPortfolioProject(ctx, id)
}

func (s *service) DeletePortfolioProject(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePortfolioProject(ctx, id)
}

func (s *service) ListPackages(ctx context.Context, query ListQuery) ([]domain.Package, int64, error) {
	return s.repo.ListPackages(ctx, query)
}

func (s *service) GetPackage(ctx context.Context, id uuid.UUID) (*domain.Package, error) {
	return s.repo.FindPackage(ctx, id)
}

func (s *service) GetPackageBySlug(ctx context.Context, slug string, activeOnly bool) (*domain.Package, error) {
	return s.repo.FindPackageBySlug(ctx, slug, activeOnly)
}

func (s *service) CreatePackage(ctx context.Context, in CreatePackageInput) (*domain.Package, error) {
	slug, err := s.resolveSlug(ctx, &domain.Package{}, in.Slug, in.NameEn, in.NameAr, uuid.Nil)
	if err != nil {
		return nil, err
	}
	projectTypeID, err := s.resolveProjectType(ctx, in.ProjectTypeID)
	if err != nil {
		return nil, err
	}
	price, err := parsePrice(in.PriceAmount)
	if err != nil {
		return nil, err
	}
	currency, err := s.resolveCurrency(ctx, in.CurrencyCode)
	if err != nil {
		return nil, err
	}

	pkg := &domain.Package{
		ID:            uuid.New(),
		ProjectTypeID: projectTypeID,
		Slug:          slug,
		NameAr:        strings.TrimSpace(in.NameAr),
		NameEn:        strings.TrimSpace(in.NameEn),
		DescriptionAr: in.DescriptionAr,
		DescriptionEn: in.DescriptionEn,
		PriceAmount:   price,
		CurrencyCode:  currency,
		DeliveryDays:  in.DeliveryDays,
		IsFeatured:    boolOr(in.IsFeatured, false),
		IsActive:      boolOr(in.IsActive, true),
		SortOrder:     in.SortOrder,
	}
	if err := s.repo.CreatePackage(ctx, pkg, toFeatures(in.Features)); err != nil {
		return nil, err
	}
	return s.repo.FindPackage(ctx, pkg.ID)
}

func (s *service) UpdatePackage(ctx context.Context, id uuid.UUID, in UpdatePackageInput) (*domain.Package, error) {
	fields := map[string]any{}
	if in.Slug != nil {
		slug, err := s.resolveSlug(ctx, &domain.Package{}, *in.Slug, "", "", id)
		if err != nil {
			return nil, err
		}
		fields["slug"] = slug
	}
	if in.ProjectTypeID != nil {
		projectTypeID, err := s.resolveProjectType(ctx, in.ProjectTypeID)
		if err != nil {
			return nil, err
		}
		fields["project_type_id"] = projectTypeID
	}
	if in.PriceAmount != nil {
		price, err := parsePrice(*in.PriceAmount)
		if err != nil {
			return nil, err
		}
		fields["price_amount"] = price
	}
	if in.CurrencyCode != nil {
		currency, err := s.resolveCurrency(ctx, *in.CurrencyCode)
		if err != nil {
			return nil, err
		}
		fields["currency_code"] = currency
	}
	setString(fields, "name_ar", in.NameAr)
	setString(fields, "name_en", in.NameEn)
	setPointer(fields, "description_ar", in.DescriptionAr)
	setPointer(fields, "description_en", in.DescriptionEn)
	if in.DeliveryDays != nil {
		fields["delivery_days"] = *in.DeliveryDays
	}
	if in.IsFeatured != nil {
		fields["is_featured"] = *in.IsFeatured
	}
	if in.IsActive != nil {
		fields["is_active"] = *in.IsActive
	}
	if in.SortOrder != nil {
		fields["sort_order"] = *in.SortOrder
	}

	var features *[]domain.PackageFeature
	if in.Features != nil {
		converted := toFeatures(*in.Features)
		features = &converted
	}

	if err := s.repo.UpdatePackage(ctx, id, fields, features); err != nil {
		return nil, err
	}
	return s.repo.FindPackage(ctx, id)
}

func (s *service) DeletePackage(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePackage(ctx, id)
}

// resolveSlug falls back to the English name, then the Arabic one, and refuses
// a slug already taken by another row of the same table.
func (s *service) resolveSlug(ctx context.Context, model any, provided, primary, fallback string, excludeID uuid.UUID) (string, error) {
	slug := Slugify(provided)
	if slug == "" {
		slug = Slugify(primary)
	}
	if slug == "" {
		slug = Slugify(fallback)
	}
	if slug == "" {
		return "", apperr.Validation("a slug is required").WithFields(apperr.FieldError{
			Field:   "slug",
			Message: "could not be generated from the provided name",
		})
	}

	exists, err := s.repo.SlugExists(ctx, model, slug, excludeID)
	if err != nil {
		return "", err
	}
	if exists {
		return "", apperr.Conflict("the slug " + slug + " is already in use")
	}
	return slug, nil
}

func (s *service) resolveProjectType(ctx context.Context, raw *string) (*uuid.UUID, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, apperr.Validation("invalid project type").WithFields(apperr.FieldError{
			Field:   "project_type_id",
			Message: "must be a valid uuid",
		})
	}
	if _, err := s.repo.FindProjectType(ctx, id); err != nil {
		return nil, err
	}
	return &id, nil
}

func (s *service) resolveCurrency(ctx context.Context, code string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	exists, err := s.repo.CurrencyExists(ctx, normalized)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", apperr.Validation("unsupported currency").WithFields(apperr.FieldError{
			Field:   "currency_code",
			Message: normalized + " is not an active currency",
		})
	}
	return normalized, nil
}

func parsePrice(raw string) (decimal.Decimal, error) {
	price, err := decimal.NewFromString(strings.TrimSpace(raw))
	if err != nil {
		return decimal.Decimal{}, apperr.Validation("invalid price").WithFields(apperr.FieldError{
			Field:   "price_amount",
			Message: "must be a decimal number",
		})
	}
	if price.IsNegative() {
		return decimal.Decimal{}, apperr.Validation("invalid price").WithFields(apperr.FieldError{
			Field:   "price_amount",
			Message: "must not be negative",
		})
	}
	return price.Round(2), nil
}

func parseDate(raw *string) (*time.Time, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.DateOnly, strings.TrimSpace(*raw))
	if err != nil {
		return nil, apperr.Validation("invalid date").WithFields(apperr.FieldError{
			Field:   "completed_at",
			Message: "must use the YYYY-MM-DD format",
		})
	}
	return &parsed, nil
}

func toImages(inputs []PortfolioImageInput) []domain.PortfolioProjectImage {
	images := make([]domain.PortfolioProjectImage, 0, len(inputs))
	for i, input := range inputs {
		sortOrder := input.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		images = append(images, domain.PortfolioProjectImage{
			ObjectKey: input.ObjectKey,
			AltAr:     input.AltAr,
			AltEn:     input.AltEn,
			SortOrder: sortOrder,
		})
	}
	return images
}

func toFeatures(inputs []PackageFeatureInput) []domain.PackageFeature {
	features := make([]domain.PackageFeature, 0, len(inputs))
	for i, input := range inputs {
		sortOrder := input.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		features = append(features, domain.PackageFeature{
			TextAr:    input.TextAr,
			TextEn:    input.TextEn,
			SortOrder: sortOrder,
		})
	}
	return features
}

func boolOr(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func setString(fields map[string]any, column string, value *string) {
	if value != nil {
		fields[column] = strings.TrimSpace(*value)
	}
}

func setPointer(fields map[string]any, column string, value *string) {
	if value != nil {
		fields[column] = value
	}
}
