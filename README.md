# Receipt Processor

Receipt Processor is a lightweight API backend built in Go that processes retail receipts to calculate reward points based on a set of business rules. The project is designed using an n‑tier architecture with separate layers for API handling, business logic, persistence, and middleware.

## What This Project Does

- **Process Receipts:** Accepts a JSON payload of receipt data, validates it, and calculates reward points according to specific rules (e.g., points per alphanumeric character in the retailer name, bonus points for round totals, etc.).
- **Retrieve Points:** Provides an endpoint to look up the points awarded for a processed receipt via its unique ID.
- **Duplicate Prevention:** Uses a hash of the receipt data to prevent storing duplicate receipts.
- **Rate Limiting:** Implements a sliding window rate limiter (using Redis) to throttle incoming requests.
- **Logging with Context:** All logs include a unique request ID, making it easier to trace requests through the system.

## Running the Application Using Docker

This project includes a Dockerfile and a Docker Compose configuration to simplify deployment. The Docker setup creates two services:

- **receipt_processor:** The Go API service.
- **redis:** A Redis service used for the rate limiter.

To run the application in Docker, use the following Makefile command:

```bash
make docker-run
```

This command will:

1. Tear down any existing Docker Compose stack (removing images if necessary).
2. Build and start the containers in detached mode.
3. You can then access the API at `http://localhost:8080`

*Warning:* Make sure both ports (8080 and 6379) are not in use.

## Running Tests

The project includes a comprehensive set of unit and integration tests for the API, middleware, repository, and service layers. To run all tests, use:

```bash
make test
```

For a test coverage report (which generates a `coverage.out` file and an HTML report), run:

```sh
make test-coverage
```

## Makefile Commands

The provided Makefile includes the following targets:

1. **test**

    Runs all tests in the project:

    ```sh
    make test
    ```

2. **Coverage**

    Runs tests with coverage and generates a coverage HTML report:

    ```sh
    make test-coverage
    ```

3. **Run**

    Builds and runs the Docker containers using Docker Compose:

    ```sh
    make docker-run
    ```
