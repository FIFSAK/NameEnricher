# Name Enricher API

A Go-based REST API that enriches personal data by automatically determining age, gender, and nationality based on a person's name.

## Features

- CRUD operations for persons, genders, and nationalities
- Automatic enrichment of personal data using external APIs
  1. Age - https://api.agify.io/?name=Dmitriy
  2. Gender - https://api.genderize.io/?name=Dmitriy
  3. Nationality - https://api.nationalize.io/?name=Dmitriy
- PostgreSQL database storage
- Swagger documentation

## Getting Started

### Prerequisites

- Go 1.18+
- PostgreSQL database

### Installation

1. Clone the repository

2. Create a `.env` file with database configuration like [.env.example](.env.example):
```bash
```env
 DATABASE_DSN=postgres://postgres:pass@localhost:5432/postgres?sslmode=disable PORT=8080
```
3. Run the application:
```bash
go run cmd/main.go
```
4. Access Swagger documentation:
   http://localhost:8080/swagger/index.html
## Database Schema

The application uses three tables:
- `persons`: Stores personal information
- `genders`: Reference table for gender types
- `nationalities`: Reference table for nationality codes
