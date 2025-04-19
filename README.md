# Go Crafting API & Calculator Backend

[![Go Version](https://img.shields.io/github/go-mod/go-version/dubbie/crafting-api.svg)](https://golang.org/dl/)

<!-- Add other badges later: build status, code coverage, license -->

This project is a RESTful API built with Go, designed to manage crafting recipes, items, and related entities for a game or application. It serves as a backend for applications needing crafting data and potentially includes logic for calculating crafting requirements.

This project emphasizes Go best practices, including:

- **Layered Architecture:** Clear separation between handlers (HTTP), services (business logic), and storage (data access).
- **Dependency Injection:** Dependencies are explicitly injected via constructors, avoiding globals.
- **Interfaces:** Used extensively for decoupling services and storage layers, promoting testability.
- **Context Propagation:** `context.Context` is passed through layers for cancellation and deadlines.
- **Clean Error Handling:** Wrapping errors for context and mapping to appropriate HTTP statuses.
- **Generic Programming:** Utilizing Go 1.18+ generics for reusable components like pagination.

## Key Technologies

- **Language:** Go (1.18+)
- **Web Framework/Router:** [Chi (v5)](https://github.com/go-chi/chi)
- **Database:** MySQL
- **Database Interaction:** Standard `database/sql`, [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql), [sqlx](https://github.com/jmoiron/sqlx) (for simplified data handling)
- **Query Building:** [Squirrel](https://github.com/Masterminds/squirrel) (for dynamic SQL generation)
- **Migrations:** [golang-migrate/migrate](https://github.com/golang-migrate/migrate) (CLI)
- **Configuration:** [Viper](https://github.com/spf13/viper) (reading from `.env` files and environment variables)
- **Request Parsing:** [gorilla/schema](https://github.com/gorilla/schema) (for query parameters)

## Project Structure

```bash
crafting_api/
├── cmd/
│ └── server/
│ └── main.go # Application entry point, wiring
├── internal/
│ ├── app/ # Application-specific reusable components
│ │ └── pagination/
│ │ └── pagination.go # Generic pagination types and helpers
│ ├── config/
│ │ └── config.go # Configuration loading (Viper)
│ ├── domain/ # Core data structures (models)
│ │ └── item.go
│ │ └── ... # Other models (recipe.go, etc.)
│ ├── handler/ # HTTP request handlers (chi)
│ │ ├── item_handler.go
│ │ ├── routes.go # Router setup
│ │ └── ... # Other handlers
│ ├── platform/ # Infrastructure-level code
│ │ └── database/
│ │ └── database.go # Database connection setup (sqlx)
│ ├── service/ # Business logic layer
│ │ ├── item_service.go
│ │ ├── item_service_impl.go
│ │ ├── list_service.go # Generic list service/handler factory
│ │ └── ... # Other services
│ └── storage/ # Data persistence layer
│ ├── errors.go # Storage-specific errors (ErrNotFound)
│ ├── item_storage.go # Item storage interface
│ ├── mysql/ # MySQL implementation of storage interfaces
│ │ └── item_store.go
│ │ └── ... # Other store implementations
│ └── ... # Other storage interfaces
├── migrations/ # Database migration files (.sql)
│ ├── 000001_create_initial_tables.down.sql
│ └── 000001_create_initial_tables.up.sql
├── pkg/ # (Optional) Libraries safe for external use
├── go.mod # Go module definition
├── go.sum # Dependency checksums
├── .env # Local environment variables (add to .gitignore!)
├── .env.example # Example environment variables
└── .gitignore # Git ignore rules
```

## Prerequisites

- Go 1.18+
- MySQL Server (running)
- `migrate` CLI tool ([Installation Guide](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate))

## Setup & Installation

1.  **Clone the repository:**
    ```bash
    git clone <your-repo-url>
    cd crafting_api
    ```
2.  **Install Go dependencies:**
    ```bash
    go mod tidy
    ```
3.  **Configure Environment:**
    - Copy the example environment file: `cp .env.example .env`
    - Edit the `.env` file with your actual database credentials and desired server port.
    ```dotenv
    # .env
    SERVER_PORT=8080
    DB_HOST=127.0.0.1
    DB_PORT=3306
    DB_USER=your_db_user
    DB_PASSWORD=your_secret_password
    DB_NAME=crafting_db
    ```
    **Important:** Ensure the database specified in `DB_NAME` exists on your MySQL server. You might need to create it manually (`CREATE DATABASE crafting_db;`).

## Running Migrations

Use the `migrate` CLI tool to apply database schema changes. Run the command from the project root directory.

```bash
# Replace USER, PASS, HOST, PORT, DBNAME with your actual values from .env
migrate -database 'mysql://USER:PASS@tcp(HOST:PORT)/DBNAME' -path migrations up

# Example using typical local dev values:
# migrate -database 'mysql://root:your_secret_password@tcp(127.0.0.1:3306)/crafting_db' -path migrations up
```

To revert the last migration:

```
migrate -database 'mysql://USER:PASS@tcp(HOST:PORT)/DBNAME' -path migrations down 1
```

## Running the Application

Once the database is migrated, you can run the API server:

```bash
go run ./cmd/server/main.go
```

The server will start, typically on http://localhost:8080 (or the port specified in your .env).
API Endpoints (Implemented)

    GET /health: Health check endpoint. Returns 200 OK.

    GET /v1/items: Lists items with pagination, filtering, and sorting.

        Query Parameters:

            page (int, default: 1)

            per_page (int, default: 15, max: 100)

            sort (string, e.g., name_asc, created_at_desc)

            name (string, filters by name, uses LIKE %name%)

            is_raw_material (bool, e.g., true or false)

    GET /v1/items/{itemID}: Retrieves a single item by its numeric ID. Returns 200 OK or 404 Not Found.
