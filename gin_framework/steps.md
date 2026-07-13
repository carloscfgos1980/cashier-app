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

## 4. Get change route
1. TransactionRequest represents the request body for a transaction.
2. ChangeLine represents a line in the change response.
3. auxiliary fuctions for the response format
- formatEuroAmount
- formatChangeLine
4. calculateChange calculates the change to be given based on the available bills.
4.1 Create a slice to hold the change bills to be returned.
4.2 Sort the bills in descending order of denomination to give the largest bills first. This is a save precaution, in this case unnecessary because I already have that condiction in the query
4.3 Iterate over the sorted bills and calculate how many of each denomination can be given as change.
4.3.1 If the change amount is zero or less, break out of the loop.
4.3.2 If the bill value is less than or equal to the change amount and there are bills available, calculate how many can be given.
4.3.3 Append the calculated number of bills to the changeBills slice.
4.3.5 Subtract the value of the given bills from the change amount.
4.4 If there is still change left to be given, return an error indicating insufficient funds.
4.5 Return the calculated change bills.
5. GetChangeHandler handles the change calculation and updates the bill inventory accordingly.
5.1 Return a handler function that processes the request to calculate change and update the bill inventory.
5.2 Parse the request body into a TransactionRequest struct
5.3 Validate that the amount paid is not less than the amount due
5.4 Calculate the change amount in cents to avoid floating point precision issues
5.5 Get the total amount of bills in the register to check if there are sufficient funds for change.
5.6 Check if there are sufficient funds in the register for change.
5.7 Get the available bills from the database.
5.8 Calculate the change to be given.
5.9 Persist the dispensed change so the current bill inventory stays in sync.
5.9.1 Retrieve the bill from the database to get its current quantity.
5.9.2 Calculate the new quantity after dispensing the change.
5.9.3 Update the bill quantity in the database.
5.10 Format the change response.
5.11 Return the formatted change response as JSON.
6. main. Define the API routes and their corresponding handlers
	router.POST("/api/change", handlers.GetChangeHandler(cfg))

## 5. Serve client
	router.StaticFile("/", "./client/index.html")
	router.StaticFile("/index.html", "./client/index.html")
	router.StaticFile("/app.js", "./client/app.js")
	router.StaticFile("/styles.css", "./client/styles.css")

## 6. TestGetChangeRoute_Integration
1. Check for the TEST_DB_URL environment variable, fallback to DB_URL if not set.
2. Connect to the database.
3. Create a new database queries instance.
4. Clear the bills table before running the test to ensure a clean state.
5. Insert test bills into the database.
6. Create a config instance with the database queries.
7. Set Gin to test mode and create a new router for testing.
8. Create a new Gin router and register the GetChangeHandler route.
9. Register the GetChangeHandler route with the router.
10. Prepare the request body for the change request.
11. Create a new HTTP request to the /api/change endpoint with the prepared body.
12. Serve the HTTP request using the router and record the response.
13. Unmarshal the response body into a slice of ChangeLine structs.
14. Check the first change line for the 5 euro bill.
15. Check the second change line for the 2 euro bill.

## README
- Description
- Features
- Tech Stack
- Project Structure
- Prerequisites
- Environment Variables
- Database Setup
- Run the App
- Available Routes
- Test