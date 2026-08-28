package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/midocss/website/internal/domain"
	"github.com/midocss/website/pkg/apperr"
)

// fakeRepository keeps just enough state to exercise the service rules without
// a database: taken slugs, known project types and active currencies.
type fakeRepository struct {
	Repository

	takenSlugs   map[string]bool
	projectTypes map[uuid.UUID]*domain.ProjectType
	currencies   map[string]bool

	createdPackage   *domain.Package
	createdFeatures  []domain.PackageFeature
	createdProject   *domain.PortfolioProject
	createdImages    []domain.PortfolioProjectImage
	updatedFields    map[string]any
	replacedFeatures *[]domain.PackageFeature
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		takenSlugs:   map[string]bool{},
		projectTypes: map[uuid.UUID]*domain.ProjectType{},
		currencies:   map[string]bool{"IQD": true, "USD": true},
	}
}

func (f *fakeRepository) SlugExists(_ context.Context, _ any, slug string, _ uuid.UUID) (bool, error) {
	return f.takenSlugs[slug], nil
}

func (f *fakeRepository) CurrencyExists(_ context.Context, code string) (bool, error) {
	return f.currencies[code], nil
}

func (f *fakeRepository) FindProjectType(_ context.Context, id uuid.UUID) (*domain.ProjectType, error) {
	projectType, ok := f.projectTypes[id]
	if !ok {
		return nil, apperr.NotFound("project type not found")
	}
	return projectType, nil
}

func (f *fakeRepository) CreateProjectType(_ context.Context, projectType *domain.ProjectType) error {
	f.projectTypes[projectType.ID] = projectType
	f.takenSlugs[projectType.Slug] = true
	return nil
}

func (f *fakeRepository) CreatePackage(_ context.Context, pkg *domain.Package, features []domain.PackageFeature) error {
	f.createdPackage = pkg
	f.createdFeatures = features
	return nil
}

func (f *fakeRepository) FindPackage(_ context.Context, _ uuid.UUID) (*domain.Package, error) {
	if f.createdPackage == nil {
		return nil, apperr.NotFound("package not found")
	}
	return f.createdPackage, nil
}

func (f *fakeRepository) UpdatePackage(_ context.Context, _ uuid.UUID, fields map[string]any, features *[]domain.PackageFeature) error {
	f.updatedFields = fields
	f.replacedFeatures = features
	return nil
}

func (f *fakeRepository) CreatePortfolioProject(_ context.Context, project *domain.PortfolioProject, images []domain.PortfolioProjectImage) error {
	f.createdProject = project
	f.createdImages = images
	return nil
}

func (f *fakeRepository) FindPortfolioProject(_ context.Context, _ uuid.UUID) (*domain.PortfolioProject, error) {
	if f.createdProject == nil {
		return nil, apperr.NotFound("portfolio project not found")
	}
	return f.createdProject, nil
}

func assertCode(t *testing.T, err error, want apperr.Code) {
	t.Helper()
	var appErr *apperr.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected an *apperr.Error, got %v", err)
	}
	if appErr.Code != want {
		t.Fatalf("error code = %q, want %q", appErr.Code, want)
	}
}

func TestCreateProjectTypeGeneratesSlug(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)

	created, err := svc.CreateProjectType(context.Background(), CreateProjectTypeInput{
		NameAr:   "متجر إلكتروني",
		NameEn:   "E-Commerce Store",
		ColorHex: "#AABBCC",
	})
	if err != nil {
		t.Fatalf("CreateProjectType: %v", err)
	}
	if created.Slug != "e-commerce-store" {
		t.Errorf("slug = %q, want %q", created.Slug, "e-commerce-store")
	}
	if created.ColorHex != "#aabbcc" {
		t.Errorf("color = %q, want it normalized to lowercase", created.ColorHex)
	}
	if !created.IsActive {
		t.Error("project types should default to active")
	}
}

func TestCreateProjectTypeRejectsDuplicateSlug(t *testing.T) {
	repo := newFakeRepository()
	repo.takenSlugs["landing-page"] = true
	svc := NewService(repo)

	_, err := svc.CreateProjectType(context.Background(), CreateProjectTypeInput{
		NameAr:   "صفحة هبوط",
		NameEn:   "Landing Page",
		ColorHex: "#123456",
	})
	assertCode(t, err, apperr.CodeConflict)
}

func TestCreatePackageValidatesPriceAndCurrency(t *testing.T) {
	svc := NewService(newFakeRepository())
	ctx := context.Background()

	base := CreatePackageInput{
		NameAr:       "الباقة الأساسية",
		NameEn:       "Basic Package",
		PriceAmount:  "250000",
		CurrencyCode: "IQD",
	}

	tests := []struct {
		name  string
		price string
		code  string
	}{
		{name: "negative price", price: "-1", code: "IQD"},
		{name: "malformed price", price: "not-a-number", code: "IQD"},
		{name: "unknown currency", price: "250000", code: "eur"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := base
			in.PriceAmount = tc.price
			in.CurrencyCode = tc.code

			_, err := svc.CreatePackage(ctx, in)
			if err == nil {
				t.Fatal("expected an error")
			}
			assertCode(t, err, apperr.CodeValidation)
		})
	}
}

func TestCreatePackageNormalizesCurrencyAndFeatureOrder(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)

	created, err := svc.CreatePackage(context.Background(), CreatePackageInput{
		NameAr:       "الباقة الاحترافية",
		NameEn:       "Pro Package",
		PriceAmount:  "1250.555",
		CurrencyCode: "usd",
		Features: []PackageFeatureInput{
			{TextAr: "تصميم", TextEn: "Design"},
			{TextAr: "استضافة", TextEn: "Hosting"},
		},
	})
	if err != nil {
		t.Fatalf("CreatePackage: %v", err)
	}
	if created.CurrencyCode != "USD" {
		t.Errorf("currency = %q, want USD", created.CurrencyCode)
	}
	if got := created.PriceAmount.String(); got != "1250.56" {
		t.Errorf("price = %s, want it rounded to 1250.56", got)
	}
	if len(repo.createdFeatures) != 2 || repo.createdFeatures[1].SortOrder != 1 {
		t.Errorf("features = %+v, want the list order preserved", repo.createdFeatures)
	}
}

func TestCreatePackageRejectsUnknownProjectType(t *testing.T) {
	svc := NewService(newFakeRepository())
	missing := uuid.New().String()

	_, err := svc.CreatePackage(context.Background(), CreatePackageInput{
		NameAr:        "باقة",
		NameEn:        "Package",
		PriceAmount:   "10",
		CurrencyCode:  "IQD",
		ProjectTypeID: &missing,
	})
	assertCode(t, err, apperr.CodeNotFound)
}

func TestUpdatePackageOnlyTouchesProvidedFields(t *testing.T) {
	repo := newFakeRepository()
	repo.createdPackage = &domain.Package{ID: uuid.New()}
	svc := NewService(repo)

	name := "Renamed"
	if _, err := svc.UpdatePackage(context.Background(), repo.createdPackage.ID, UpdatePackageInput{NameEn: &name}); err != nil {
		t.Fatalf("UpdatePackage: %v", err)
	}

	if len(repo.updatedFields) != 1 || repo.updatedFields["name_en"] != "Renamed" {
		t.Errorf("updated fields = %+v, want only name_en", repo.updatedFields)
	}
	// A nil Features pointer must leave the existing features alone.
	if repo.replacedFeatures != nil {
		t.Error("features were replaced even though the payload omitted them")
	}
}

func TestCreatePortfolioProjectDefaultsToUnpublished(t *testing.T) {
	repo := newFakeRepository()
	svc := NewService(repo)

	if _, err := svc.CreatePortfolioProject(context.Background(), CreatePortfolioProjectInput{
		TitleAr: "موقع شركة",
		TitleEn: "Company Website",
		Images:  []PortfolioImageInput{{ObjectKey: "a.png"}, {ObjectKey: "b.png"}},
	}); err != nil {
		t.Fatalf("CreatePortfolioProject: %v", err)
	}

	if repo.createdProject.IsPublished {
		t.Error("new portfolio projects must stay unpublished until reviewed")
	}
	if repo.createdProject.Slug != "company-website" {
		t.Errorf("slug = %q, want company-website", repo.createdProject.Slug)
	}
	if len(repo.createdImages) != 2 {
		t.Errorf("images = %d, want 2", len(repo.createdImages))
	}
}

func TestCreatePortfolioProjectRejectsInvalidDate(t *testing.T) {
	svc := NewService(newFakeRepository())
	invalid := "2026-13-45"

	_, err := svc.CreatePortfolioProject(context.Background(), CreatePortfolioProjectInput{
		TitleAr:     "موقع",
		TitleEn:     "Website",
		CompletedAt: &invalid,
	})
	assertCode(t, err, apperr.CodeValidation)
}

func TestPublicQueryForcesActiveOnly(t *testing.T) {
	query := ListQuery{Page: 2, PerPage: 5}
	if query.ActiveOnly() {
		t.Error("dashboard queries must see inactive rows")
	}

	public := query.Public()
	if !public.ActiveOnly() {
		t.Error("public queries must be restricted to published/active rows")
	}
	page, perPage := public.Normalized()
	if page != 2 || perPage != 5 {
		t.Errorf("pagination = %d/%d, want 2/5", page, perPage)
	}
}
