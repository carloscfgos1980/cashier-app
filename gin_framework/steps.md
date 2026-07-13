# STEPS

## 1. Set up

1. Start the appp
2. Copy files ad directories from previos version:
- sqlc.yaml
- .gitignore
- .env
- sql directory
- internal/config
- client

3. Run sqlc generate 
4. main
-  Load configuration from environment variables
- Connect to the database
- Create a new database queries instance
- Assign the database queries instance to the config
- Initialize the Gin router
- Set trusted proxies to nil to avoid warnings in Gin 1.7+
-  Define a simple health check route
- Start the server on the specified port