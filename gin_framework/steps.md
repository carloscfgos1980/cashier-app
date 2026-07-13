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

## Get bills route
1. Define a struct to format the response
2. GetBillsHandler retrieves all bills from the database and returns them in a formatted response.
2.1 Return a handler function that retrieves bills from the database and formats the response.
2.2 Use cfg.DB to access the database and retrieve bills
2.3 Create a slice to hold the formatted bills for the response
2.4 Format each bill and append it to the response slice
2.5 Return the formatted bills as a JSON response
3.  Define the API routes and their corresponding handlers
	router.GET("/api/bills", handlers.GetBillsHandler(cfg))