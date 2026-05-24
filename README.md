# Project Manager API

A robust, fully CRUD and REST-compliant backend API for the Project Manager application, built with **Go** (Golang), **GORM** (object-relational mapping), and **MySQL**.

---

## Repository Architecture & Integration

This repository is one of two parts that make up the Project Manager application:

- **Backend API (Go)**: [project-manager-api](https://github.com/QuentinLReinette/project-manager-api) - Serves the REST API and handles data persistence.
- **Frontend Client (React)**: [project-manager-web](https://github.com/QuentinLReinette/project-manager-web) - Renders the collaborative Kanban user interface.

These two repositories communicate via HTTP REST endpoints to form a decoupled client-server architecture.

---

## Technical Stack & Architecture

- **Language:** Go (1.26+)
- **Database:** MySQL 8.0 (relational persistence)
- **ORM:** GORM (Go Object Relational Mapping)
- **Authentication:** JWT (JSON Web Tokens) & Bcrypt password hashing
- **Concurrency & Cancellation:** Request context propagation (`context.Context`) through to GORM queries for execution safety.
- **Database Optimizations:** GORM connection pooling (Max Open/Idle connections config), optimized access-control queries, and database transactions for atomicity.
- **Security & Middleware:** JWT authentication gate, request logging, and CORS middleware (handling options preflights for frontend integration).
- **Development Tooling:** Air (hot-reload inside Docker)
- **Containerization:** Docker & Docker Compose

---

## Prerequisites

To run this project, you need the following tools installed:

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)

---

## Installation & Running

1. **Clone the repository:**

   ```bash
   git clone https://github.com/QuentinLReinette/project-manager-api.git
   cd project-manager-api
   ```

2. **Configure Environment Variables:**

   A `.env.example` file is provided in the repository. Create a `.env` file at the root and fill in the required environment variables for database connection and JWT secret:

   ```bash
   cp .env.example .env
   ```

3. **Start the API and Database Containers:**

   Start the services with Docker Compose:

   ```bash
   docker compose up -d --build
   ```

   This command starts:
   - `mysql_db` (service mapping port `3306` inside the private network)
   - `api` (service running Go's Air utility for hot-reloads, mapped to port `8080` on the host)

   _Note: Database migrations run automatically via GORM on startup. No manual database seeding is required. You can register your first account directly from the frontend UI or by sending a `POST` request to `/api/auth/register`._

4. **Verify running status:**
   Check the status of the containers:

   ```bash
   docker compose ps
   ```

   You can also ping the API sanity check endpoint:

   ```bash
   curl http://localhost:8080/ping
   ```

   Expected response:

   ```json
   { "message": "pong", "status": "running" }
   ```

---

## API Endpoints Documentation

All endpoints are fully documented in a separate file for readability. You can find detailed descriptions of request payloads, URL parameters, error codes, and JSON responses in `API_DOCUMENTATION.md`.

For a quick reference of the available endpoints, see below:

- **Auth**: Registration (`POST /api/auth/register`), Login (`POST /api/auth/login`), Profile Info (`GET /api/auth/me`), Logout (`POST /api/auth/logout`)
- **Users**: Search (`GET /api/users?q={query}`)
- **Projects**: List (`GET /api/projects`), Create (`POST /api/projects`), Detail (`GET /api/projects/{id}`), Update (`PUT /api/projects/{id}`), Delete (`DELETE /api/projects/{id}`), Add Participant (`POST /api/projects/{id}/participants`), Remove Participant (`DELETE /api/projects/{id}/participants`), Leave (`POST /api/projects/{id}/leave`)
- **Tasks**: List (`GET /api/tasks?project_id={id}`), Create (`POST /api/tasks`), Detail (`GET /api/tasks/{id}`), Update (`PUT /api/tasks/{id}`), Delete (`DELETE /api/tasks/{id}`)
