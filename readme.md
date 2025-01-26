# Gopher Social Backend 🚀

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Description 📚

Gopher Social Backend is the API server for Gopher Social, a social media platform built for Gophers (Go enthusiasts!). This backend provides a robust and scalable foundation for a social networking experience, offering features for user management, content creation, social interactions, and more.

This API is built using Go, leveraging the Gin Gonic framework for high performance and efficiency. It's designed with best practices in mind, including:

*   **Authentication & Authorization:** Secure JWT-based authentication and role-based authorization to protect your data and ensure only authorized users can perform specific actions.
*   **Scalability:** Designed to handle a growing user base and increasing data volume with PostgreSQL and Redis integration.
*   **Observability:** Comprehensive logging with Logrus, request tracing with Request IDs, and health check endpoints for monitoring and maintenance.
*   **Developer Experience:**  Well-documented API using Swagger, making it easy for frontend and mobile developers to integrate with the backend.

## Features ✅

*   **User Authentication:**
    *   User Registration with Email Verification
    *   Login and Logout
    *   Password Reset (Forgot Password Flow)
    *   Account Activation and Resend Activation Link
*   **User Profile Management:**
    *   Update Profile Information (First Name, Last Name, Website, Social Links)
    *   Retrieve Own Profile and User Profiles by Identifier
*   **Social Interactions:**
    *   Follow and Unfollow Users
    *   Get Followers and Following Lists for Users
*   **Post Management:**
    *   Create, Update, and Delete Posts
    *   Retrieve Posts by ID
    *   List Posts for Logged-in User and by User Identifier
*   **Post Likes & Dislikes:**
    *   Like and Unlike Posts
    *   Dislike and Undislike Posts
    *   List Liked and Disliked Posts for Logged-in User and by User Identifier
*   **Comment Management:**
    *   Create, Update, and Delete Comments on Posts
    *   Retrieve Comments by ID
    *   List Comments for Logged-in User and by User Identifier for a Post
*   **Comment Likes & Dislikes:**
    *   Like and Unlike Comments
    *   Dislike and Undislike Comments
    *   List Liked and Disliked Comments for a Post by Logged-in User and by User Identifier
*   **News Feed:**
    *   Retrieve Latest Posts for a Personalized Feed
    *   Get a Specific Post with its Comments
*   **Moderation & Administration Actions:**
    *   Timeout Users
    *   Remove User Timeout
    *   List Timed Out Users
    *   Deactivate and Activate Users
    *   Ban and Unban Users
    *   Delete Comments and Posts (Moderator/Admin Roles)
*   **Health Checks:**
    *   Router Health
    *   Redis Health
    *   PostgreSQL Health
*   **Middleware & Enhancements:**
    *   Request Rate Limiting (using Redis)
    *   Request Timeout Handling
    *   CORS (Cross-Origin Resource Sharing) Support
    *   Request Logging with Request IDs and Real IP detection
    *   Panic Recovery

## Technologies Used 🛠️

*   [Go](https://go.dev/) - Programming Language
*   [Gin Gonic](https://gin-gonic.com/) - Web Framework
*   [PostgreSQL](https://www.postgresql.org/) - Relational Database
*   [Redis](https://redis.io/) - In-memory Data Store (Caching, Rate Limiting)
*   [pgx](https://github.com/jackc/pgx) - PostgreSQL Driver for Go
*   [go-redis/redis](https://github.com/redis/go-redis) - Redis client for Go
*   [logrus](https://github.com/sirupsen/logrus) - Structured Logger
*   [golang-jwt/jwt](https://github.com/golang-jwt/jwt/v5) - JSON Web Tokens (JWT) for Authentication
*   [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) - Password Hashing
*   [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/) - Containerization
*   [Swaggo](https://github.com/swaggo/swag) - Swagger for API Documentation
*   [Make](https://www.gnu.org/software/make/) - Build Automation Tool

## Environment Variables ⚙️

Configuration is managed through environment variables. Example environment files are provided in the `.envs` directory:

*   `.envs/.postgres.env.example`: PostgreSQL database configuration.
*   `.envs/.server.env.example`: Server and application settings.

**Required Environment Variables:**

*   `SERVER_MODE`:  Set to `release` for production, defaults to `release`.
*   `SERVER_PORT`:  Port for the server to listen on, defaults to `:8080`.
*   `POSTGRES_HOST`: PostgreSQL host address, defaults to `localhost`.
*   `POSTGRES_PORT`: PostgreSQL port, defaults to `5432`.
*   `POSTGRES_USER`: PostgreSQL username, defaults to `postgres`.
*   `POSTGRES_PASSWORD`: PostgreSQL password, defaults to `postgres`.
*   `POSTGRES_DBNAME`: PostgreSQL database name, defaults to `gopher`.
*   `POSTGRES_SSLMODE`: PostgreSQL SSL mode, defaults to `disable`.
*   `REDIS_ADDR`: Redis server address, defaults to `localhost:6379`.
*   `REDIS_PASSWORD`: Redis password (if any), defaults to empty.
*   `REDIS_DB`: Redis database number, defaults to `0`.
*   `JWT_ACCESS_SECRET`: Secret key for JWT access tokens.
*   `JWT_REFRESH_SECRET`: Secret key for JWT refresh tokens.
*   `JWT_RESET_SECRET`: Secret key for JWT password reset tokens.
*   `JWT_ACTIVATION_SECRET`: Secret key for JWT activation tokens.
*   `DATABASE_URL`: Database connection URL, if using URL configuration.
*   `DOMAIN`: Base domain URL for activation and password reset links, defaults to `http://localhost:8080`.

Refer to the example files for more details and other optional configurations.

## Docker Setup 🐳

To get started with Docker, ensure you have Docker and Docker Compose installed.

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd gopher-social-backend
    ```

2.  **Copy environment files and configure them:**
    ```bash
    cp .envs/.postgres.env.example .envs/.postgres.env
    cp .envs/.server.env.example .envs/.server.env
    # ... and modify the .env files with your desired settings ...
    ```

3.  **Build and start the Docker containers:**
    ```bash
    make docker-build
    ```
    This command will build the Docker image and start the services defined in `docker-compose.yml` in detached mode.

4.  **Access the API:**
    The API will be accessible at `http://localhost:8080`.

5.  **Stop and clean up Docker containers:**
    ```bash
    make docker-clean
    ```
    This command will stop and remove the containers, images, volumes, and network associated with the project.

## Makefile Commands 🛠️

The `Makefile` provides convenient commands for development and deployment:

*   **`make help`**:  Displays a help message listing available commands.
*   **`make docker-build`**: Builds and starts the Docker containers in detached mode.
*   **`make docker-clean`**: Stops and cleans up Docker resources (containers, images, volumes, and build cache).
*   **`make gen-docs`**: Generates Swagger API documentation.
*   **`make migrate-create`**: Creates a new database migration file.
*   **`make migrate-up`**: Applies all pending database migrations.
*   **`make migrate-down`**: Rolls back database migrations (prompts for the number of steps).
*   **`make migrate-clean`**: Drops all database objects created by migrations.

## API Documentation 📚

The API documentation is generated using Swagger and is available at:

[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

After starting the server, you can access the Swagger UI to explore the API endpoints, models, and try out requests.

To regenerate the documentation after making changes to the codebase, run:
```bash
make gen-docs
```

## Health Check Script 🩺

The `healthCheck.sh` script is used by Docker to verify the health of the application. It performs HTTP GET requests to the health check endpoints:

* `/api/v1/health/router`
* `/api/v1/health/redis`
* `/api/v1/health/postgres`

If all checks pass, the script exits with code 0, otherwise with code 1, indicating an unhealthy state.

## License 📝

This project is licensed under the MIT License - see the [LICENSE](license) file for details.

## Contributing 🤝

Contributions are welcome! Please feel free to submit pull requests, report issues, or suggest new features to improve Gopher Social Backend.

**Author:** Rohit Vilas Ingole ([rohit.vilas.ingole@gmail.com](mailto:rohit.vilas.ingole@gmail.com))
