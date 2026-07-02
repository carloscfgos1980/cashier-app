# Steps: Money change app

## 1. Set up

### start the app

```bach 
go mod init github.com/carloscfgos1980/cashier-app
```

### Packages needed
github.com/joho/godotenv 
github.com/lib/pq
github.com/stretchr/testify
github.com/google/uuid

### Copy fiules to set up:
- .gitignore
- sql.yaml
- json.go
- .env

### main

1. Load environment variables from .env file
2. Get configuration from environment variables
3. Get the port from environment variables, default to 8080 if not set
4. Connect to the database
5. variable for the apiConfig struct
6. Set up the HTTP server and routes
7. Start the HTTP server
8. Listen and serve

### health route
1. handlerHealth responds with a simple health check
2. main: Define health route and its corresponding handlers

### SQL

1. Create bills table
2. Run migration
```bash
cd sql/schema
goose postgres "postgres://carlosinfante@localhost:5432/cashier?sslmode=disable" up
```
3. Create queries
3.1 CreateBill :one
3.2 GetBills
3.3 GetBillByDenomination
3.4 UpdateBil
3.5 GetBillsTotalAmount
4. run sqlc command to generate GO code from SQl
5. Update apiConfig for database queries
6. database queries variable and variable for the apiConfig struct

Note: Here I had to make some thinking cuz the app with in memory data would use a map. In this case I can use the key of the map in order to sort it out the bills decrecincly by value. With data base I need to create a table that holds the value (denomination) and its quantity.

## 2. Get bills route
1, handlerBillsGet handles the GET /api/bills endpoint. It retrieves all bills from the database and returns
1.1 Retrieve all bills from the database
1.2 Define a struct to format the response
1.3 Create a slice to hold the formatted bills for the response
1.4 Respond with the formatted bills as JSON
2. main: Define route
	mux.HandleFunc("GET /api/bills", apiCfg.handlerBillsGet)

## 3. Create or update bills

1. handlerBillsCreateUpdate handles the creation or update of bills in the database.
1.1 Decode the request body into a slice of Bill structs
1.2 Iterate over the bills and create or update them in the database
1.2.1 Validate the denomination
1.2.2 convert the denomination from float32 to int (cents) to avoid floating
1.2.3 Check if the bill already exists in the database
1.2.4 If the bill does not exist, create it; otherwise, update the quantity
1.2.5 check if the bill already exists
1.2.6 if it exists, update the quantity
1.3 response with the updated bills
2. main: Define route
	mux.HandleFunc("POST /api/bills", apiCfg.handlerBillsCreateUpdate)

## 4. Change route

### structs
- Change represents the request body for the change calculation.
- Bill represents a bill denomination and its quantity.
- ChangeLine represents a line in the change response.

### helpers functions
- formatChangeResponse formats the change bills into a slice of ChangeLine for the response.
- formatChangeLine formats a single line of change information.
- formatEuroAmount formats a float64 value as a Euro currency string.
- calculateChange calculates the change to be given based on the available bills.

### handlerGetChange handles the GET /change endpoint to calculate and return the change.
1. Decode the request body into a Change struct.
2. Validate the input amounts.
3. Calculate the change amount.
4. Get the total amount of bills in the store to ensure we have enough change.
5. Check if we have enough change in the store.
6. Convert changeAmount to cents to avoid floating point precision issues.
7. Get the available bills from the database.
8. Calculate the change to be given using the available bills.
9. If changeBills is empty, it means we couldn't make the exact change with the available bills.
10. update the quantities of the bills in the database
10.1 Get the current bill from the database by denomination.
10.2 Calculate the new quantity of the bill after giving change.
10.3 Update the bill quantity in the database.
11. Respond with the formatted change response.

### main
	mux.HandleFunc("POST /api/change", apiCfg.handlerGetChange)

## 5. client
I just ask pilot to build a client. I had to write several queries to get the desired UI 

## 6. unit testing
