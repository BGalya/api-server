Q1

1. The original JWT claims didn’t include a user’s unique ID, making it hard to confirm who owns what.
So I updated the Claims struct - added the unique ID and also stored it the Login func. 


2. The balance-related functions (getBalance, depositBalance, and withdrawBalance) had the above issues:
- Did not verify that the requester had the "user" permission. This allowed admins to access these endpoints.
  To fix this I added a role-based check.
- Trusted the userId in the query string, so it could allow access to a user who did not owned the account.
  To fix this, I compared the query string with the logged user`s userId.
- I also checked if the input is valid in the getBalance function.

Q2

This setup (in main func) creates a basic server that listens to requests.

Q3

Initially, I thought of using a decorator pattern (similar to Python's @ decorator syntax) to log requests and responses.
However, since Go does not support decorators in the same way as Python, I decided to manually call the logging functions
where needed.
I created a separate package (logger.go) that holds all the necessary functions for logging.
This package contains the core logic for writing logs to the access.log file.
I wrote a WriteToFile function that handles writing the request and response data to the file in JSON format.
This function is called after processing each request in the relevant API handler functions (Register, Login, 
AccountsHandler, BalanceHandler).

