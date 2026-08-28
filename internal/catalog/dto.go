package catalog

// Inputs are shared by the dashboard endpoints. Every optional field is a
// pointer so a PATCH can tell "leave untouched" apart from "set to empty".
type CreateProjectTypeInput struct {
	Slug          string  `json:"slug" binding:"omitempty,max=96"`
	NameAr        string  `json:"name_ar" binding:"required,max=128"`
	NameEn        string  `json:"name_en" binding:"required,max=128"`
	DescriptionAr *string `json:"description_ar"`
	DescriptionEn *string `json:"description_en"`
	ColorHex      string  `json:"color_hex" binding:"required,hexcolor"`
	IconObjectKey *string `json:"icon_object_key"`
	SortOrder     int     `json:"sort_order"`
	IsActive      *bool   `json:"is_active"`
}

type UpdateProjectTypeInput struct {
	Slug          *string `json:"slug" binding:"omitempty,max=96"`
	NameAr        *string `json:"name_ar" binding:"omitempty,max=128"`
	NameEn        *string `json:"name_en" binding:"omitempty,max=128"`
	DescriptionAr *string `json:"description_ar"`
	DescriptionEn *string `json:"description_en"`
	ColorHex      *string `json:"color_hex" binding:"omitempty,hexcolor"`
	IconObjectKey *string `json:"icon_object_key"`
	SortOrder     *int    `json:"sort_order"`
	IsActive      *bool   `json:"is_active"`
}

type PortfolioImageInput struct {
	ObjectKey string  `json:"object_key" binding:"required"`
	AltAr     *string `json:"alt_ar" binding:"omitempty,max=200"`
	AltEn     *string `json:"alt_en" binding:"omitempty,max=200"`
	SortOrder int     `json:"sort_order"`
}

type CreatePortfolioProjectInput struct {
	ProjectTypeID  *string               `json:"project_type_id" binding:"omitempty,uuid4"`
	Slug           string                `json:"slug" binding:"omitempty,max=160"`
	TitleAr        string                `json:"title_ar" binding:"required,max=200"`
	TitleEn        string                `json:"title_en" binding:"required,max=200"`
	DescriptionAr  *string               `json:"description_ar"`
	DescriptionEn  *string               `json:"description_en"`
	ExternalURL    *string               `json:"external_url" binding:"omitempty,url"`
	CoverObjectKey *string               `json:"cover_object_key"`
	CompletedAt    *string               `json:"completed_at" binding:"omitempty,datetime=2006-01-02"`
	SortOrder      int                   `json:"sort_order"`
	IsPublished    *bool                 `json:"is_published"`
	Images         []PortfolioImageInput `json:"images" binding:"omitempty,dive"`
}

type UpdatePortfolioProjectInput struct {
	ProjectTypeID  *string `json:"project_type_id" binding:"omitempty,uuid4"`
	Slug           *string `json:"slug" binding:"omitempty,max=160"`
	TitleAr        *string `json:"title_ar" binding:"omitempty,max=200"`
	TitleEn        *string `json:"title_en" binding:"omitempty,max=200"`
	DescriptionAr  *string `json:"description_ar"`
	DescriptionEn  *string `json:"description_en"`
	ExternalURL    *string `json:"external_url" binding:"omitempty,url"`
	CoverObjectKey *string `json:"cover_object_key"`
	CompletedAt    *string `json:"completed_at" binding:"omitempty,datetime=2006-01-02"`
	SortOrder      *int    `json:"sort_order"`
	IsPublished    *bool   `json:"is_published"`
	// Images replaces the whole gallery when present; omit it to keep it as is.
	Images *[]PortfolioImageInput `json:"images" binding:"omitempty,dive"`
}

type PackageFeatureInput struct {
	TextAr    string `json:"text_ar" binding:"required,max=255"`
	TextEn    string `json:"text_en" binding:"required,max=255"`
	SortOrder int    `json:"sort_order"`
}

type CreatePackageInput struct {
	ProjectTypeID *string               `json:"project_type_id" binding:"omitempty,uuid4"`
	Slug          string                `json:"slug" binding:"omitempty,max=160"`
	NameAr        string                `json:"name_ar" binding:"required,max=160"`
	NameEn        string                `json:"name_en" binding:"required,max=160"`
	DescriptionAr *string               `json:"description_ar"`
	DescriptionEn *string               `json:"description_en"`
	PriceAmount   string                `json:"price_amount" binding:"required"`
	CurrencyCode  string                `json:"currency_code" binding:"required,len=3"`
	DeliveryDays  *int                  `json:"delivery_days" binding:"omitempty,min=1"`
	IsFeatured    *bool                 `json:"is_featured"`
	IsActive      *bool                 `json:"is_active"`
	SortOrder     int                   `json:"sort_order"`
	Features      []PackageFeatureInput `json:"features" binding:"omitempty,dive"`
}

type UpdatePackageInput struct {
	ProjectTypeID *string `json:"project_type_id" binding:"omitempty,uuid4"`
	Slug          *string `json:"slug" binding:"omitempty,max=160"`
	NameAr        *string `json:"name_ar" binding:"omitempty,max=160"`
	NameEn        *string `json:"name_en" binding:"omitempty,max=160"`
	DescriptionAr *string `json:"description_ar"`
	DescriptionEn *string `json:"description_en"`
	PriceAmount   *string `json:"price_amount"`
	CurrencyCode  *string `json:"currency_code" binding:"omitempty,len=3"`
	DeliveryDays  *int    `json:"delivery_days" binding:"omitempty,min=1"`
	IsFeatured    *bool   `json:"is_featured"`
	IsActive      *bool   `json:"is_active"`
	SortOrder     *int    `json:"sort_order"`
	// Features replaces the whole list when present; omit it to keep it as is.
	Features *[]PackageFeatureInput `json:"features" binding:"omitempty,dive"`
}

// ListQuery drives both the dashboard listings and the public ones; the public
// routes force PublishedOnly/ActiveOnly regardless of what the client sends.
type ListQuery struct {
	Page            int    `form:"page" binding:"omitempty,min=1"`
	PerPage         int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search          string `form:"search" binding:"omitempty,max=160"`
	ProjectTypeSlug string `form:"type" binding:"omitempty,max=96"`
	Featured        *bool  `form:"featured"`

	activeOnly bool
}

func (q ListQuery) Normalized() (page, perPage int) {
	page, perPage = q.Page, q.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	return page, perPage
}

// Public returns a copy of the query restricted to published/active rows.
func (q ListQuery) Public() ListQuery {
	q.activeOnly = true
	return q
}

func (q ListQuery) ActiveOnly() bool { return q.activeOnly }
