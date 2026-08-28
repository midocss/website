package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Currency struct {
	Code         string          `gorm:"primaryKey;size:3" json:"code"`
	NameAr       string          `gorm:"size:64" json:"name_ar"`
	NameEn       string          `gorm:"size:64" json:"name_en"`
	Symbol       string          `gorm:"size:8" json:"symbol"`
	ExchangeRate decimal.Decimal `gorm:"type:numeric(18,6)" json:"exchange_rate"`
	IsDefault    bool            `json:"is_default"`
	IsActive     bool            `json:"is_active"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

func (Currency) TableName() string { return "currencies" }

// ProjectType classifies both portfolio entries and packages (e-commerce site,
// landing page, ...). Its colour and icon drive the public UI.
type ProjectType struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Slug          string         `gorm:"size:96;uniqueIndex" json:"slug"`
	NameAr        string         `gorm:"size:128" json:"name_ar"`
	NameEn        string         `gorm:"size:128" json:"name_en"`
	DescriptionAr *string        `json:"description_ar,omitempty"`
	DescriptionEn *string        `json:"description_en,omitempty"`
	ColorHex      string         `gorm:"size:7" json:"color_hex"`
	IconObjectKey *string        `json:"icon_object_key,omitempty"`
	SortOrder     int            `json:"sort_order"`
	IsActive      bool           `json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProjectType) TableName() string { return "project_types" }

type PortfolioProject struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectTypeID  *uuid.UUID     `gorm:"type:uuid" json:"project_type_id,omitempty"`
	Slug           string         `gorm:"size:160;uniqueIndex" json:"slug"`
	TitleAr        string         `gorm:"size:200" json:"title_ar"`
	TitleEn        string         `gorm:"size:200" json:"title_en"`
	DescriptionAr  *string        `json:"description_ar,omitempty"`
	DescriptionEn  *string        `json:"description_en,omitempty"`
	ExternalURL    *string        `gorm:"column:external_url" json:"external_url,omitempty"`
	CoverObjectKey *string        `json:"cover_object_key,omitempty"`
	CompletedAt    *time.Time     `gorm:"type:date" json:"completed_at,omitempty"`
	SortOrder      int            `json:"sort_order"`
	IsPublished    bool           `json:"is_published"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	ProjectType *ProjectType            `gorm:"foreignKey:ProjectTypeID" json:"project_type,omitempty"`
	Images      []PortfolioProjectImage `gorm:"foreignKey:PortfolioProjectID" json:"images,omitempty"`
}

func (PortfolioProject) TableName() string { return "portfolio_projects" }

type PortfolioProjectImage struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PortfolioProjectID uuid.UUID `gorm:"type:uuid" json:"portfolio_project_id"`
	ObjectKey          string    `json:"object_key"`
	AltAr              *string   `gorm:"size:200" json:"alt_ar,omitempty"`
	AltEn              *string   `gorm:"size:200" json:"alt_en,omitempty"`
	SortOrder          int       `json:"sort_order"`
	CreatedAt          time.Time `json:"created_at"`
}

func (PortfolioProjectImage) TableName() string { return "portfolio_project_images" }

// Package is a fixed-price service offering. Prices are decimals paired with an
// explicit currency code so no rounding happens in Go floats.
type Package struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectTypeID *uuid.UUID      `gorm:"type:uuid" json:"project_type_id,omitempty"`
	Slug          string          `gorm:"size:160;uniqueIndex" json:"slug"`
	NameAr        string          `gorm:"size:160" json:"name_ar"`
	NameEn        string          `gorm:"size:160" json:"name_en"`
	DescriptionAr *string         `json:"description_ar,omitempty"`
	DescriptionEn *string         `json:"description_en,omitempty"`
	PriceAmount   decimal.Decimal `gorm:"type:numeric(14,2)" json:"price_amount"`
	CurrencyCode  string          `gorm:"size:3" json:"currency_code"`
	DeliveryDays  *int            `json:"delivery_days,omitempty"`
	IsFeatured    bool            `json:"is_featured"`
	IsActive      bool            `json:"is_active"`
	SortOrder     int             `json:"sort_order"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index" json:"-"`

	ProjectType *ProjectType     `gorm:"foreignKey:ProjectTypeID" json:"project_type,omitempty"`
	Features    []PackageFeature `gorm:"foreignKey:PackageID" json:"features,omitempty"`
}

func (Package) TableName() string { return "packages" }

type PackageFeature struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PackageID uuid.UUID `gorm:"type:uuid" json:"package_id"`
	TextAr    string    `gorm:"size:255" json:"text_ar"`
	TextEn    string    `gorm:"size:255" json:"text_en"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

func (PackageFeature) TableName() string { return "package_features" }
