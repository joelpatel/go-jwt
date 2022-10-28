# go-jwt

## Functionality

- Store hashed password in database.
- Use jwt to issue token and refresh token.
- Production project structure.
- Validation on incoming signup requests.

## Instructions

Add following details in the `.env` file at root directory of this project. [NOTE: do not change the spelling or case]

- PORT
- MONGODB_URL
- MONGODB_DB
- JWT_PRIVATE_KEY
- PASSWORD_HASH_COST

Run `go mod tidy` to install dependencies.  
Execute `go run ./...` to start the server.

Download and import the postman requests from `postman` directory.  
Modify and run those sample APIs.
