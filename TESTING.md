# 🧪 Testing Guide for Joyo Abadi

This document provides comprehensive information about testing in the Joyo Abadi application.

## Test Structure

```
├── models/
│   ├── user_test.go           # User model tests
│   ├── product_test.go        # Product model tests
│   └── transaction_test.go    # Transaction model tests
├── utils/
│   ├── session_test.go        # Session management tests
│   ├── locals_test.go         # Fiber locals tests
│   ├── logger_test.go         # Logging tests
│   └── benchmark_test.go      # Performance benchmarks
├── middleware/
│   ├── auth_test.go           # Authentication middleware tests
│   └── rate_limiter_test.go   # Rate limiting tests
├── controllers/
│   ├── auth_test.go           # Authentication controller tests (complex)
│   ├── auth_simple_test.go    # Authentication controller tests (simple)
│   └── home_test.go           # Home controller tests
├── integration_test.go        # End-to-end integration tests
├── run-tests.sh              # Comprehensive test runner
└── Makefile                  # Test automation commands
```

## Test Categories

### 1. Unit Tests
Test individual components in isolation:
- **Models**: Database operations, validations, relationships
- **Utils**: Session management, logging, helper functions
- **Middleware**: Authentication, rate limiting, request processing
- **Controllers**: HTTP handlers, business logic

### 2. Integration Tests
Test complete workflows:
- Authentication flow (register → login → access protected routes)
- Product management flow
- Error handling across the application
- Rate limiting behavior

### 3. Performance Tests
Benchmark critical operations:
- Password hashing and comparison
- Session operations
- Logging performance
- Database queries

## Running Tests

### Quick Start
```bash
# Run all tests
make test

# Run specific test categories
make test-unit
make test-integration
make test-coverage
```

### Detailed Test Execution
```bash
# Run comprehensive test suite
./run-tests.sh

# Run tests for specific package
go test -v ./models/
go test -v ./utils/
go test -v ./middleware/
go test -v ./controllers/

# Run with coverage
go test -v -cover ./...

# Run with race detection
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

## Test Features

### ✅ What's Tested

#### Models
- User creation and validation
- Email uniqueness constraints
- Password hashing verification
- Product CRUD operations
- Product-User relationships
- Transaction creation and queries
- Database constraints and validations

#### Utils
- Session initialization and configuration
- Session data storage and retrieval
- User ID management in sessions
- Logger initialization with different levels
- Fiber request logging middleware
- Local variables management

#### Middleware
- Authentication middleware behavior
- Unauthorized access handling
- Rate limiting functionality
- IP-based rate limiting
- Rate limit reset behavior

#### Controllers
- Login with valid/invalid credentials
- User registration flow
- Duplicate email handling
- Password hashing during registration
- Template rendering (basic checks)

#### Integration
- Complete authentication workflow
- Protected route access control
- Health endpoint functionality
- Error handling (404, 405, etc.)
- Rate limiting integration

### 🔧 Test Utilities

#### Database Setup
```go
func setupTestDB() *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    // Auto-migrate schemas for testing
    db.AutoMigrate(&models.User{}, &models.Product{}, &models.Transaction{})
    return db
}
```

#### Application Setup
```go
func setupTestApp() *fiber.App {
    utils.InitLogger()
    utils.InitSession()
    app := fiber.New()
    return app
}
```

## Test Configuration

### Environment Variables
Tests use in-memory SQLite databases and don't require external dependencies.

### Test Data
- Isolated test databases for each test
- Predictable test data setup
- Clean state for each test run

## Coverage Reports

### Generating Coverage
```bash
# Generate HTML coverage report
make test-coverage

# View coverage in browser
open coverage.html
```

### Coverage Goals
- **Models**: >90% (critical business logic)
- **Utils**: >85% (core utilities)
- **Middleware**: >80% (security components)
- **Controllers**: >75% (HTTP handlers)
- **Overall**: >80%

## Performance Testing

### Benchmarks Available
- Password hashing operations
- Session management
- Logging performance
- String operations

### Running Benchmarks
```bash
# Run all benchmarks
make test-bench

# Run specific benchmarks
go test -bench=BenchmarkPasswordHashing ./utils/
go test -bench=BenchmarkSessionOperations ./utils/
```

## Continuous Integration

### Pre-commit Checks
```bash
# Format, lint, and test
make check
```

### Test Automation
The `run-tests.sh` script provides:
- Comprehensive test execution
- Coverage reporting
- Race condition detection
- Static analysis
- Build verification

## Best Practices

### Writing Tests
1. **Arrange-Act-Assert** pattern
2. **Descriptive test names** that explain the scenario
3. **Independent tests** that don't rely on each other
4. **Clean up** resources after tests
5. **Test both success and failure cases**

### Test Organization
1. **Group related tests** using `t.Run()`
2. **Use setup/teardown functions** for common operations
3. **Mock external dependencies** when necessary
4. **Keep tests simple** and focused

### Example Test Structure
```go
func TestUserModel(t *testing.T) {
    db := setupTestDB()
    
    t.Run("Create User", func(t *testing.T) {
        // Arrange
        user := models.User{Email: "test@example.com", Password: "password"}
        
        // Act
        result := db.Create(&user)
        
        // Assert
        assert.NoError(t, result.Error)
        assert.NotZero(t, user.ID)
    })
}
```

## Troubleshooting

### Common Issues
1. **Template not found errors**: Expected in controller tests without proper template setup
2. **Session errors**: Ensure `utils.InitSession()` is called before session tests
3. **Database errors**: Check that test database is properly migrated
4. **Race conditions**: Use `-race` flag to detect concurrent access issues

### Debug Tips
1. Use `t.Log()` for debugging test output
2. Run individual tests with `-v` flag for verbose output
3. Use `go test -run TestSpecificTest` to run single tests
4. Check test coverage to identify untested code paths

## Future Improvements

### Planned Enhancements
- [ ] API endpoint testing with JSON responses
- [ ] Database transaction testing
- [ ] File upload testing
- [ ] Email sending mock testing
- [ ] WebSocket testing (if implemented)
- [ ] Load testing scenarios

### Tools to Consider
- [ ] Testify mock for complex mocking
- [ ] GoConvey for BDD-style testing
- [ ] Ginkgo for advanced test organization
- [ ] Docker for integration test environments
