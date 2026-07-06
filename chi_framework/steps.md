# CASHIER CHI FRAMEWORK

## 1. Set up

1. Start the app
mod init github.com/carloscfgos1980/cashier-app

2. Copy auxiliary files and folfers fromprevious version
- .env
- .gitignore
- sql directory
- sqlc.yaml
- internal/env
- internal/json

3. Run sqlc command to generate Go code from SQL 

4. api.go
4.1 Structs
- application holds the application-wide dependencies for the HTTP server
- config holds the configuration for the application
- dbConfig holds the database configuration for the application
4.2 mount sets up the routes and middleware for the HTTP server
4.3 run starts the HTTP server

5. main
5.1 Load environment variables from .env file
5.2 create a context
5.3 initialize logger
5.4 database connection
5.5 create the application
5.6 run the application