
# Golang RESTful API

This project is a RESTful API implemented in Go (Golang) using MongoDB as the database. It covers basic CRUD operations for user management and is designed with Go modules and scalable architecture in mind.

## Table of Contents

- [Requirements](#requirements)
- [Project Structure](#project-structure)
- [Setup Instructions](#setup-instructions)
- [Running the Application](#running-the-application)
- [API Endpoints](#api-endpoints)
- [Environment Variables](#environment-variables)
- [CI/CD Pipeline](#cicd-pipeline)
- [Security Scans](#security-scans)
- [Testing](#testing)
- [Future Work](#future-work)


## Requirements

- Go v1.20 or higher
- MongoDB instance
- Git


## Project Structure

```plaintext
golang-restful-api/
├── .github/
│   └── workflows/
│       └── ci_cd_pipeline.yml
├── db/
│   ├── connect.go
│   └── connect_test.go
├── handlers/
│   ├── user.go
│   └── user_test.go
├── models/
│   └── user.go
├── scripts/
│   └── create-eb-environment.sh
├── .env
├── .gitignore
├── go.mod
├── go.sum
├── main.go
├── main_test.go
├── option-settings.json
└── README.md
```

- `main.go`: Entry point of the application.
- `.github/workflows/`: Contains the CI/CD pipeline configuration using GitHub Actions.
- `db/`: Manages database connections, with corresponding tests in connect_test.go.
- `handlers/`: Contains the handler functions for API endpoints with corresponding tests in user_test.go.
- `models/`: Defines the data models for the application.
- `scripts/`: Contains automation scripts for deployment; create-eb-environment.sh.
- `.env`: Stores the environment variables for the application.
- `option-settings.json`: Holds the configuration settings for the Elastic Beanstalk environment.


## Setup Instructions

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/lep13/golang-restful-api.git
   cd golang-restful-api
   ```

2. **Install Dependencies**:

   Ensure you have Go installed and then install dependencies using:

   ```bash
   go mod tidy
   ```

3. **Set Up Environment Variables**:

   Create an `.env` file in the root directory:

   ```plaintext
   MONGO_URI=your_mongodb_uri
   PORT=5000
   ```

4. **Run the Application**:

   Start the server by running:

   ```bash
   go run main.go
   ```

5. **Test API Endpoints**:

   You can use tools like Postman or curl to interact with the API endpoints listed below.


## API Endpoints

| Method | Endpoint        | Description               |
|--------|-----------------|---------------------------|
| POST   | /users          | Create a new user         |
| GET    | /users          | Retrieve all users        |
| GET    | /users/{id}     | Retrieve a specific user  |
| DELETE | /users/{id}     | Delete a specific user    |
| PUT    | /users/{id}     | Update a specific user    |
| GET	   | /health	      | Health check of the API   |


## Environment Variables

Ensure you have the following variables set in your `.env` file:
- `MONGO_URI`: The connection string for your MongoDB instance.
- `PORT`: Port on which the server will run (default: 5000).


## CI/CD Pipeline

The CI/CD pipeline is configured using GitHub Actions. It includes the following stages:
- Build: Compiles the Go application.
- Security Scan: Uses Gosec for Go code security scanning and Trivy for vulnerability scans.
- Deployment: Deploys the application to AWS Elastic Beanstalk.
- Health Check: Ensures the environment health is stable after deployment.
- CloudWatch Logs: Automatically enables CloudWatch logs for monitoring.
- Automatic Rollback: Rolls back the environment if deployment fails.


## Security Scans

Security scanning is implemented with:
- Gosec: Scans the Go codebase for potential security issues.
- Trivy: Scans for vulnerabilities in dependencies and application files.
- Snyk: (Optional) Can be integrated for further dependency scanning.


## Testing

Test cases have been implemented for all directories including handlers, models, and database connections. Use the following command to run the tests:
   ```bash
   go test ./...
   ```

## Future Work

- Implement user authentication and authorization.
- Integrate with Docker for containerized deployment.
- Enhance the CI/CD pipeline with better monitoring and alerting mechanisms.
