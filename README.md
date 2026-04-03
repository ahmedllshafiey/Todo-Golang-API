# Todo API (Golang & Gin) ocumentation

A high-performance RESTful API for managing tasks, built with **Go (Golang)** and the **Gin Gonic** framework. This project implements secure JWT authentication and robust PostgreSQL integration using connection pooling.

---

## 🛠 Features

- **JWT Authentication**: Secure user registration and login flows.
- **PostgreSQL Support**: Efficient database interactions using `pgxpool`.
- **Middleware Protection**: Dedicated auth middleware to secure sensitive routes.
- **CRUD Operations**: Full Create, Read, Update, and Delete capabilities for Todo items.
- **Health Monitoring**: Root endpoint to verify API and Database connectivity.

---

## 🏗 Project Structure

The application follows a modular internal architecture:

```

internal/config       # Environment variable and configuration management
internal/database     # PostgreSQL connection logic using pgx
internal/handlers     # HTTP logic for authentication and todo management
internal/middleware   # Authorization logic (JWT verification)

```

---

## 🚀 Getting Started

### 1. Prerequisites

- **Go**: 1.21 or higher
- **PostgreSQL**: A running instance (local or Docker)
- **Git**

### 2. Installation & Setup

1. Clone the repository:

```bash
git clone <your-repository-url>
cd todo-api
```

2. Install dependencies:

```bash
go mod tidy
```

3. Create rename `.env_model` to a `.env` file in the root and fill required information:

```env
PORT=3000
DATABASE_URL=postgres://user:password@localhost:5432/dbname
JWT_SECRET=your_secret_key
```

4. Run the application:

```bash
go run cmd/main.go
```

---

## 📡 API Endpoints

### Public Endpoints

| Method | Endpoint       | Description              |
| ------ | -------------- | ------------------------ |
| GET    | /              | Health check & DB status |
| POST   | /auth/register | Register a new user      |
| POST   | /auth/login    | Login and receive JWT    |

### Protected Endpoints (Authorization: Bearer `<token>`)

| Method | Endpoint        | Description                    |
| ------ | --------------- | ------------------------------ |
| GET    | /todos          | Get all user todos             |
| GET    | /todos/:id      | Get a specific todo            |
| POST   | /todos          | Create a new todo              |
| PUT    | /todos/:id      | Update a todo (title/status)   |
| DELETE | /todos/:id      | Delete a todo                  |
| GET    | /protected-test | Verify middleware connectivity |

---

## 💻 Code Overview (`main.go`)

The entry point initializes the system by loading configurations, establishing a database pool, and setting up the Gin router with injected dependencies:

```go
func main() {
    // Configuration & DB initialization
    cfg, _ := config.Load()
    pool, _ := database.Connect(cfg.DatabaseURL)
    defer pool.Close()

    router := gin.Default()

    // Routes
    router.POST("/auth/register", handlers.CreateUserHandler(pool))
    router.POST("/auth/login", handlers.LoginHandler(pool, cfg))

    protected := router.Group("/todos")
    protected.Use(middleware.AuthMiddleWare(cfg))
    {
        protected.POST("", handlers.CreateTodoHandler(pool))
        protected.GET("", handlers.GetAllTodosHandler(pool))
        protected.GET("/:id", handlers.GetTodoByIDHandler(pool))
        protected.PUT("/:id", handlers.UpdateTodoByIDHandler(pool))
        protected.DELETE("/:id", handlers.DeleteTodoHandler(pool))
    }

    router.Run(":" + cfg.Port)
}
```

---

## 🧪 Testing

A `.http` file is included for testing with the REST Client extension:

1. Use `/auth/register` to create a user.
2. Use `/auth/login` to get a JWT token.
3. Replace the Authorization header in your requests with:

```
Authorization: Bearer YOUR_TOKEN_HERE
```
