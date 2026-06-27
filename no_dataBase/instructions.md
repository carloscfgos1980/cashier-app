# CASHIER INSTRUCTIONS

## CalculateChange
1. Validate input amounts
2. Validate that the amount paid is sufficient
3. Calculate the change in cents
4. Create a map to hold the change distribution
5. Convert the cashier map to a slice of denominations in cents for easier processing
6. Create a map to convert back from cents to float64 denominations
7. Populate the denominations slice and the conversion map
8. Sort the denominations in descending order to prioritize larger bills and coins
9. Iterate through the sorted denominations to calculate the change distribution
10. If the remaining change is zero or less, we can break early
11. value of the denomination in float64 for logging and map access
12. If the cashier has this denomination and the change left is greater than or equal to the denomination, we can use it
13. Calculate the maximum number of this denomination that can be used. Ensure we don't use more than what the cashier has
14. Update the change distribution and the remaining change
15. If there is still change left to give, it means the cashier doesn't have enough denominations to provide the exact change
16. Return the calculated change distribution

## main
1. Example usage of the CalculateChange function
2. Cashier's available denominations and their counts
3. Calculate the change to be given
4. Print Change calculation summary
5. Separate bills and coins for better readability
6. Sort bills and coins in descending order for display
7. Print the change distribution
8. Print subtotals and grand total
9. Print the remaining cashier inventory after the transaction