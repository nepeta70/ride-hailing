---
name: repository-standards
description: Enforces project-specific coding, database, and testing standards. Use this skill whenever the user asks to write new code, create database migrations, or generate unit tests for the current Go repository.
---

# Repository Standards & Coding Guidelines

This skill ensures all generated code and database schemas align with the specific architecture of this repository.

## General & Coding Rules
- **Naming Integrity:** Never rename existing entities (structs, variables, tables, fields) unless a refactor is explicitly requested.
- **Error Handling:** All errors must be categorized. Refer to `domain/errors.go` for the appropriate error types and factory functions.

## 🗄 Database Standards
- **Column Types:** Never use `TEXT` columns in database schemas or migrations. Use `VARCHAR(n)` or the most appropriate modern data type for the specific database engine.

## Unit Testing Protocol
Follow these strict requirements for all Go unit tests:

### 1. Structure & Package
- Always use the `folder_test` package naming convention (e.g., if the code is in `package auth`, the test must be `package auth_test`).
- Use a dot import to reference the package under test: `import . "github.com/nepeta70/ride-hailing/internal/path/to/package"`.

### 2. Implementation
- **Assertions:** Use the `assert` package from `"github.com/stretchr/testify/assert"`.
- **Methodology:** Use **Table-Driven Tests** for every method. Each test should iterate through a slice of anonymous structs defining inputs and expected outputs.

### 3. Mocking Strategy
- **Simple Structs:** Do NOT mock simple data structs; use real instances.
- **Interface Mocks:** - Before creating a new mock for interfaces in `internal/pkg`, check `internal/pkg/mocks/` to see if it already exists.
    - If a mock is missing, create it in the shared `internal/pkg/mocks/` directory so it can be reused across the project.

## Quality Checklist
Before finalizing any code, verify:
1. Is the test in a `_test` package?
2. Are errors categorized via `domain/errors.go`?
3. Did I avoid using `TEXT` in the DB schema?
4. Are the tests table-driven?