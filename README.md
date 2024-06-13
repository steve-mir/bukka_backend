# Bukka Backend

Welcome to the Bukka Backend project! This repository contains the backend services for the Bukka application.

## Table of Contents

- [Introduction](#introduction)
- [Technologies Used](#technologies-used)
- [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
  - [Authentication](#authentication)
    - [Register](#register)
    - [Login](#login)
    - [Profile](#profile)
- [Contributing](#contributing)
- [License](#license)

## Introduction

Bukka Backend is designed to provide robust and secure authentication services for the Bukka application. This documentation will guide you through the installation and usage of the current authentication endpoints.

## Technologies Used

- **Programming Language:** Go (Golang)
- **Framework:** Gin
- **Database:** PostgreSQL
- **Authentication:** Paseto
- **Other Tools:** Docker, Redis, Kubernetes

## Installation

1. Clone the repository:
    ```bash
    git clone https://github.com/frankoe-dev/bukka_backend.git
    cd bukka_backend
    ```

2. Install dependencies:
    ```bash
    go mod tidy
    ```

3. Set up environment variables:
    Create a `.env` file in the root directory and add the following:
    ```env
    DATABASE_URL=your_database_url
    SECRET_KEY=your_secret_key
    ```

4. Run database migrations:
    ```bash
    make migrateup
    ```

5. Start the server:
    ```bash
    go run main.go
    ```

## Usage

To use the authentication endpoints, you can use tools like [Postman](https://www.postman.com/) or [cURL](https://curl.se/). Below are the details of the available endpoints.

## API Endpoints

### Authentication

#### Register

- **Endpoint:** `/api/auth/register`
- **Method:** `POST`
- **Description:** Create a new user account.
- **Request Body:**
    ```json
    {
        "username": "your_username",
        "email": "your_email",
        "password": "your_password"
    }
    ```
- **Response:**
    ```json
    {
        "message": "User registered successfully",
        "user": {
            "id": "user_id",
            "username": "your_username",
            "email": "your_email"
        }
    }
    ```

#### Login

- **Endpoint:** `/api/auth/login`
- **Method:** `POST`
- **Description:** Authenticate a user and return a JWT token.
- **Request Body:**
    ```json
    {
        "email": "your_email",
        "password": "your_password"
    }
    ```
- **Response:**
    ```json
    {
        "message": "Login successful",
        "token": "your_jwt_token"
    }
    ```

#### Profile

- **Endpoint:** `/api/auth/profile`
- **Method:** `GET`
- **Description:** Get the profile of the authenticated user.
- **Headers:**
    ```http
    Authorization: Bearer your_jwt_token
    ```
- **Response:**
    ```json
    {
        "user": {
            "id": "user_id",
            "username": "your_username",
            "email": "your_email"
        }
    }
    ```

## Contributing

We welcome contributions from the community. Please follow these steps to contribute:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature/your-feature`).
3. Commit your changes (`git commit -m 'Add some feature'`).
4. Push to the branch (`git push origin feature/your-feature`).
5. Create a new Pull Request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
