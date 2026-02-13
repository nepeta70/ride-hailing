# AI Instructions

This file is intended to contain guidelines and rules that must always be followed by any AI assistant working in this repository. Add your instructions below:

---

## Guidelines

### General
- Never rename my entities (structs, variables, tables, fields) unless I order it explictly.

### Coding
- All errors must be categorized: check domain/errors.go

### Database
- Never use TEXT columns

### Unit tests
- should always be in package folder_test
- should use assert from "github.com/stretchr/testify/assert"
- should use table driven tests for each method tested
- refer to the struct folder using . "structbeingtestedpath"
- Don't mock simple structs

---

## Purpose

This file serves as a persistent reference for AI behavior and project-specific requirements. It should be reviewed and updated as needed to ensure compliance and alignment with project goals.