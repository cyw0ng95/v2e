# Copilot Instructions for v2e

## Documentation Guidelines

- Do **NOT** generate any documents other than `README.md`
- All project documentation should be consolidated in the `README.md` file
- Avoid creating additional markdown files, guides, or documentation files
- Keep the documentation simple and focused

## Project Guidelines

- This is a Go-based project using Go modules
- The project may contain multiple commands in the `cmd/` directory
- Follow Go best practices and conventions
- Use standard Go tooling (go build, go test, go mod)

## Testing Guidelines

- When adding new features, always consider adding test cases
- Unit tests should be written in Go using the standard `testing` package
- Integration tests should be written in Python using pytest
- Test cases should cover:
  - Normal operation paths
  - Error handling and edge cases
  - Integration between multiple services (for RPC-based features)
- Run `./build.sh -t` to execute unit tests
- Run `./build.sh -i` to execute integration tests
- Test case generation should be comprehensive and follow existing patterns in the codebase
