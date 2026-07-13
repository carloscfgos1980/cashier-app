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

## 2. Get bills route
1. Define a struct to format the response
2. GetBillsHandler retrieves all bills from the database and returns them in a formatted response.
2.1 Return a handler function that retrieves bills from the database and formats the response.
2.2 Use cfg.DB to access the database and retrieve bills
2.3 Create a slice to hold the formatted bills for the response
2.4 Format each bill and append it to the response slice
2.5 Return the formatted bills as a JSON response
3.  Define the API routes and their corresponding handlers
	router.GET("/api/bills", handlers.GetBillsHandler(cfg))

## 3. Create or update bill
1. Define a struct to handle the request for creating or updating bills
2. Auxiliary function. ValidateDenomination checks if the provided denomination is valid based on predefined denominations.
3. BillsCreateUpdateHandler handles the creation or update of bills in the database.
3.1 Return a handler function that processes the request to create or update bills.
3.2 Parse the request body into a slice of BillRequest structs
3.3 Iterate over each bill in the request and process it
3.3.1 Validate the quantity
3.3.2 alidate the denomination
3.3.3 convert the denomination from float32 to int (cents) to avoid floating point precision issues
3.3.4 Check if the bill already exists in the database
3.3.5 Bill does not exist, create a new one
3.3.6 heck if the bill already exists
3.3.7 ill exists, update the quantity
4. Return bills after processing the request
5. main. Define the API routes and their corresponding handlers
	router.POST("/api/bills", handlers.BillsCreateUpdateHandler(cfg))