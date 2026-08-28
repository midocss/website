# Backend Build Prompt (Go)

## Project Overview
I want to build the backend for a complete web platform (a single Monolith serving the admin dashboard + public website + store), for a personal (not yet a registered company) web development services business. The platform combines:
- A personal site showcasing web development services.
- A portfolio of past projects.
- A quotes system: fixed packages shown to visitors + a custom quote request form.
- A full e-commerce store: fixed service packages + additional digital products (templates, icons, tools), with a shopping cart and online payment.
- An admin dashboard with multiple permission levels to manage everything above.

## General Technical Requirements

### Language & Tools
- **Language:** Go (latest stable version).
- **Web framework:** Gin or Fiber (choose the most suitable one and explain why).
- **Database:** PostgreSQL.
- **ORM / Query Builder:** GORM or sqlx (choose whichever fits a clear, struct-first design best).
- **Migrations:** a dedicated, clear migration tool (e.g. golang-migrate).
- **Authentication:** JWT (Access Token + Refresh Token).
- **File storage:** integration with MinIO (S3-compatible object storage) for SVG icons and digital product files.
- **Architecture:** clear layering (Handler → Service → Repository), never mixing business logic with the HTTP layer.

### Project Structure
Propose a clean Go folder structure (using `/internal`, `/cmd`, `/pkg` conventions) separating:
- API layer (Handlers/Routes)
- Business logic layer (Services)
- Data access layer (Repositories)
- Models/Entities
- Configuration
- Middlewares (Auth, Logging, Error Handling, Rate Limiting)

## Functional Requirements

### 1. Users & Permissions
- Three account types:
  - **Super Admin:** full access.
  - **Staff:** limited, fully customizable permissions (e.g. a staff member can add offers but not delete them).
  - **Customers:** regular registration/login accounts to track their orders and purchases.
- Design a flexible permission system (Role-Based or Permission-Based ACL) that allows adding new roles/permissions later without major code changes.
- Endpoints: register, login, logout, refresh token, and admin-only management of users and permissions.

### 2. Project Types — Colors & Icons
- A dedicated `project_types` table containing: name, primary color (Hex), path/URL of the associated animated SVG icon (stored on MinIO), and an optional description.
- Full CRUD for project types from the dashboard (create, edit, delete, upload new SVG icon).
- Validate the color value (must be a valid Hex) and the uploaded file (SVG only, size-limited).

### 3. Portfolio
- A `portfolio_projects` table for past work: title, description, images, link (if any), and the associated project type.
- Full CRUD from the dashboard.
- A public endpoint to display the portfolio to visitors.

### 4. Quotes System
- **Fixed packages:** a `packages` table containing: name, description, price, associated project type, and a list of included features.
- **Custom quote request:** a form submitted by visitors (choose the most suitable fields: name, email, WhatsApp number, project type, request details, optional reference attachments).
- When a new request arrives: send an automatic notification (email and/or WhatsApp — design the code to be extensible for both channels, even if the actual WhatsApp API integration comes later).
- Dashboard view to review custom quote requests and update their status (new, under review, responded, rejected).

### 5. E-commerce Store
- **Products:** two types sold from the same cart:
  - Fixed service packages (linked to the packages table).
  - Digital products (templates, icons, tools) — a `digital_products` table with a file link stored on MinIO.
- **Cart:** tied to the customer's account, supports add/remove/update quantity.
- **Orders:** created on checkout, containing all cart items, the total price, and the order status.
- **Digital product delivery:** after payment confirmation, the customer gets a secure, time-limited download link (signed URL) from MinIO for the purchased files.

### 6. Online Payments
- Design a generic Payment Gateway interface so new gateways can be added later without major rework.
- **Phase 1:** integrate with **ZainCash**.
- **Future (just prepare the architecture for now):** support **SuperQi** for Visa/Mastercard.
- Track the status of every payment transaction (Pending, Success, Failed) and link it to the order.
- A webhook/callback endpoint to receive payment confirmation from the gateway.

### 7. Invoices
- After a successful payment, automatically generate a **PDF invoice** (order details, product/package breakdown, total price, customer info).
- Use a suitable Go library to generate the PDF (e.g. gofpdf, or render from an HTML template).
- Store the invoice (on MinIO or a local path) and email its link to the customer; it should also be accessible from the customer's account.

### 8. Bilingual Support (Arabic/English)
- Design the content structure to support i18n — especially for dynamic content (package names, project descriptions, project type names), not just static UI strings.
- Think through the simplest scalable approach (e.g. dual fields name_ar/name_en, or a separate translations table) and explain the pros/cons of each option.

### 9. Future Scalability
- Design the database and code so it's easy in the future to:
  - Support multiple currencies (not just Iraqi Dinar).
  - Support multiple timezones.
  - Add new payment gateways.
- No need to fully implement this now, but the current architecture must not block it.

## Non-Functional Requirements
- **Security:** password hashing (bcrypt), protection against CSRF/XSS/SQL Injection, rate limiting on sensitive endpoints (login, public forms).
- **Error handling:** a unified, consistent JSON error response format.
- **Logging:** structured logging of important requests and errors.
- **Testing:** basic unit tests for sensitive services (payments, permissions, cart).
- **API documentation:** document every endpoint (OpenAPI/Swagger preferred).

## What I Need From You Now
1. Propose the full project folder structure.
2. Design the database schema (ERD or a description of tables and relationships) based on everything above.
3. Start with the foundation: project setup (config, DB connection, router), then the users/authentication/permissions system first as the starting point.
4. After each phase, briefly explain what was completed before moving to the next phase.
