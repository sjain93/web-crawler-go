# 👨🏽‍💻 userservice
## Overview

This repository demonstrates the implementation of the Service Repository Layer Pattern in the Go programming language. The Service Repository Layer Pattern is a software architectural pattern that separates the concerns of business logic (services) from data access (repositories) in an organized and maintainable way.

## Features

- Separation of concerns: The pattern separates the business logic (services) from the data access (repositories), making the codebase more modular and easier to maintain. 
  >This modularity is enforced via Go's implementation of interfaces, and can be found as a salient feature in both the repository and service layers.
- Testability: Each layer can be unit-tested in isolation, leading to more reliable code.
- Reusability: Services and repositories can be reused in different parts of the application.
  
> Check out `main.go` for an implementation of `Echo`'s graceful shutdown feature!

Key dependencies used in this project include:
- [Echo v4](github.com/labstack/echo/v4 ) for routing and server middleware, its a great library that also supports
features like authorization and context logging
- [govalidator](github.com/asaskevich/govalidator) for HTTP request validation
- [GORM](gorm.io/gorm) for Postgres ORM
- [testify](github.com/stretchr/testify) for robust testing assertions

## Repository Structure

- `api/`: Holds the main application logic.
  - `common/`: Contains common utilities that all subdomains can use.
  - `user/`: The user subdomain.
    - `handler/`: Repository implementations for data access.
    - `model/`: Data models and structures used throughout the application.
    - `repository/`: Repository implementations for data access.
    - `service_test/`: Service tests in the "Testing Table" style.
    - `service/`: Service implementations containing business logic.
- `config/`: Configuration files for the application.
- `migrations/`: Contains the application database migration commands.
- `routes/`: Contains the server's routes.
- `main.go`: The main application file, and is the entry point for the application.

>📝 More subdomains will be added as the project expands. 

## Getting Started

#### Please note that this project was designed to be run with a Postgres database however, you may also init the project with an in-memory store to test it out.

Follow these steps to get the project up and running:

1. Clone the repository to your local machine:

   ```
   git clone https://github.com/sjain93/userservice.git
   ```

2. Navigate to the project directory:

   ```
   cd userservice
   ```

3. Install any required dependencies:

   ```
   go get -d ./...
   ```

4. Customize the configuration files in the `config/` directory according to your needs.
   > Note that this project uses `.env` files to configure the database. Check out the original `dotenv` project [here](https://github.com/motdotla/dotenv)

   Here's a sample of my `.env` file that connects to my local Postgres instance

   ```bash
    DB_USER=`YOUR USER`
    DB_PASSWORD=`YOUR PW`
    DB_NAME=`YOUR DB NAME`
    PORT=5432
    SSL_mode=disable
    ```

5. To build and run the application with Postgres setup:

   ```
   go run main.go
   ```

6. To build and run the application with an in-memory store:

   ```
   go run main.go -noDB
   ```

   The `noDB` flag will configure the repository layer to use a in memory map
   found in the `config` package.

   ```go
   type MemoryStore map[string]interface{}
   ```
## Making HTTP Requests
Postman was used in the development of this API, however, with the server running,
here are some cURL equivalents of HTTP calls that can be made:

### `POST` to `/api/users`
```cURL
curl -X POST --location 'http://localhost:8080/api/users' \
--header 'Content-Type: application/json' \
--data-raw '{
    "username": "Test User",
    "email": "testuser@gmail.com"
}'
```

### `GET` a specific user from `/api/users/{{userID}}`
```cURL
curl --location 'http://localhost:8080/api/users/{{userID}}'
```
>**Note** that the userID is currently implemented as a `MD5` hash of the `username`
and `email` fields.

---
⚙️ `OpenAPI` and `Swagger` doc spec to come!

## Usage

In this pattern, services encapsulate the business logic of your application, while repositories provide data access methods. You can use the services in your application's handlers or controllers to perform business operations.

Here's an example of how to use a service:

```go
// Import the necessary packages
import (
    "github.com/sjain93/userservice/api/user"
    "github.com/sjain93/userservice/config"
)

// Initialize the userRepository with an in memory store
memStore := config.GetInMemoryStore()
userRepository, err := user.NewUserRepository(nil, memStore)
if err != nil {
    // handle the error
}

// Initialize a service
userService := service.NewUserService(userRepository)

// Create a new user
newUser := user.User{Username: "John Doe", Email: "john@example.com"}
createdUser, err := userService.CreateUser(newUser)
if err != nil {
    // Handle the error
} else {
    // User created successfully
}
```

## Testing

To run unit tests for services and repositories, use the following command:

```
go test ./...
```
---
⚙️ More tests to come!

## Contributing

Contributions are welcome! If you'd like to contribute to this project, please follow these guidelines:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and ensure they are properly tested.
4. Submit a pull request to the `main` branch of this repository.

## License

This project is licensed under the Apache License - see the [LICENSE.md](./LICENSE.md) file for details.

---

Feel free to explore the code and adapt it to your own project's needs. Happy coding! 🚀