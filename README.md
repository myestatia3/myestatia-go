# MyEstatia

AI-Powered Real Estate CRM.

## Project Structure

*   **`myestatia-go`**: Backend (Go + Gin + Gorm + PostgreSQL). - https://github.com/myestatia/myestatia-go
*   **`myestatia`**: Frontend (React + Vite + TailwindCSS). - https://github.com/myestatia/myestatia

## Prerequisites

Ensure you have installed:
*   [Go](https://go.dev/dl/) (1.21+)
*   [Node.js](https://nodejs.org/) (18+)
*   [PostgreSQL](https://www.postgresql.org/)

---

## 1. Backend Setup (`myestatia-go`)

The backend handles business logic, database, and authentication.

### Steps:

1.  Navigate to the backend directory:
    ```bash
    cd myestatia-go
    ```

2.  Configure environment variables:
    Create a `.env` file in `myestatia-go/` with the following content (adjust your DB credentials):
    ```env
    DB_HOST=localhost
    DB_USER=postgres
    DB_PASSWORD=your_password
    DB_NAME=myestatia
    DB_PORT=5432
    DB_SSLMODE=disable
    JWT_SECRET_KEY=super_secure_secret_key
    ```

3.  Install dependencies and start the server:
    ```bash
    go mod tidy
    go run cmd/main.go
    ```
    *The server will start at `http://localhost:8080`.*

---

## 2. Frontend Setup (`myestatia`)

The frontend is a modern Chat-like user interface.

### Steps:

1.  Open a new terminal and navigate to the frontend directory:
    ```bash
    cd myestatia
    ```

2.  Install dependencies:
    ```bash
    npm install
    ```

3.  Start the development server:
    ```bash
    npm run dev
    ```
    *The web application will be available at `http://localhost:5173`.*

---

## 3. Usage

1.  Open your browser at **`http://localhost:5173`**.
2.  Click on **"Start Now"**.
3.  If it's your first time, go to the **"Create new account"** tab and register.
4.  You will automatically access the Dashboard (`/ai-actions`).
