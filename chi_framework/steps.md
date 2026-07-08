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

## 2. GetBills route

### types. go
Define a struct to format the response

### service.go
- Service defines the interface for the users service
- svc defines the struct for the users service
- NewService creates a new service for the users package
- GetBills method retrieves all bills from the database
- Add GetBills method to service interface


### handler
- handler is the HTTP handler for bills endpoints
- NewHandler creates a new handler for bills endpoints
- GetBills retrieves all bills from the database and returns them as a JSON response

### api.go
	billsService := bills.NewService(database.New(app.db), app.db)
	billsHandler := bills.NewHandler(billsService)
	r.Get("/api/bills", billsHandler.GetBills)

## 3. Create Bills or Update

### types.go
- BillRequest struct represents a bill denomination and its quantity.

### services.go
- GetBillByDenomination retrieves a bill by its denomination from the database
- CreateBill creates a new bill in the database
- UpdateBill updates the quantity of a bill in the database
- Add these 3 methods to service interfaces

### handler.go - BillsCreateUpdate handles the creation or update of bills based on the request body
1. Read the request body into a slice of BillRequest
2. Iterate over the slice of bills from the request and then Validate and process each bill in the request
2.1 Validate the denomination
2.2 convert the denomination from float32 to int (cents) to avoid floating point precision issues
2.3 Fetch the bill from the database by its denomination
2.4 bill does not exist, create it
2.5 check if the bill already exists
2.6 bill exists, update it

### api.go
	billsService := bills.NewService(database.New(app.db), app.db)
	billsHandler := bills.NewHandler(billsService)
	r.Post("/api/bills", billsHandler.BillsCreateUpdate)

## 4. Get change

### types.go
- Transaction represents the request body for the change calculation.
- ChangeLine represents a line in the change response.

### service.go
- GetBillsTotalAmount retrieves the total amount of all bills in the database
- Add "GetBillsTotalAmount" to service interface

### handler.go. GetChange

**Auxiliary fucntions**
- ValidateDenomination checks if the given denomination is valid based on predefined denominations.
- formatChangeResponse formats the change bills into a slice of ChangeLine for the response.
- formatChangeLine formats a single line of change information.
- formatEuroAmount formats a float64 value as a Euro currency string.
- calculateChange calculates the change to be given based on the available bills.

1. GetChange calculates the change to be given based on the amount paid and the amount due, and returns the change as a JSON response
1.1 Read the request body into a TransactionRequest struct
1.2 Validate the input amounts.
1.3 Calculate the change amount.
1.4 Convert the change amount to cents to avoid floating point precision issues.
1.5 Get the total amount of bills in the database.
1.6 Check if there are enough bills to give change.
1.7 Get the available bills from the database.
1.8 Calculate the change to be given using the available bills.
1.9 Calculate the change to be given using the available bills.
1.10 update the quantities of the bills in the database
1.10.1 Fetch the bill from the database by its denomination
1.10.2 calculate the new quantity of the bill after giving change
1.10.3 update the bill quantity in the database
1.11 Format the change response.
1.12 Write the change response as JSON.

### api
	billsService := bills.NewService(database.New(app.db), app.db)
	billsHandler := bills.NewHandler(billsService)
	r.Post("/api/change", billsHandler.GetChange)

Note:
Client, testing and metrics made with copilot