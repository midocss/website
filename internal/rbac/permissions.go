package rbac

// Permission slugs seeded by migration 000001. Handlers reference these
// constants instead of raw strings so a typo fails at compile time.
const (
	UsersView   = "users.view"
	UsersCreate = "users.create"
	UsersUpdate = "users.update"
	UsersDelete = "users.delete"

	RolesView   = "roles.view"
	RolesManage = "roles.manage"

	ProjectTypesView   = "project_types.view"
	ProjectTypesCreate = "project_types.create"
	ProjectTypesUpdate = "project_types.update"
	ProjectTypesDelete = "project_types.delete"

	PortfolioView   = "portfolio.view"
	PortfolioCreate = "portfolio.create"
	PortfolioUpdate = "portfolio.update"
	PortfolioDelete = "portfolio.delete"

	PackagesView   = "packages.view"
	PackagesCreate = "packages.create"
	PackagesUpdate = "packages.update"
	PackagesDelete = "packages.delete"

	QuotesView   = "quotes.view"
	QuotesUpdate = "quotes.update"
	QuotesDelete = "quotes.delete"

	ProductsView   = "products.view"
	ProductsCreate = "products.create"
	ProductsUpdate = "products.update"
	ProductsDelete = "products.delete"

	OrdersView   = "orders.view"
	OrdersUpdate = "orders.update"

	PaymentsView   = "payments.view"
	PaymentsRefund = "payments.refund"

	InvoicesView = "invoices.view"
)
