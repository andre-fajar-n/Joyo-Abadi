#!/bin/bash

# Comprehensive Test Runner for Joyo Abadi

echo "🧪 Running Comprehensive Tests for Joyo Abadi"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go first."
    exit 1
fi

print_status "Go version: $(go version)"

# Clean up any previous test artifacts
print_status "Cleaning up previous test artifacts..."
go clean -testcache

# Download dependencies
print_status "Downloading dependencies..."
go mod download
if [ $? -ne 0 ]; then
    print_error "Failed to download dependencies"
    exit 1
fi

# Run tests with different verbosity levels
run_tests() {
    local package=$1
    local description=$2
    
    print_status "Running $description..."
    
    # Run tests with coverage
    go test -v -race -coverprofile=coverage_${package//\//_}.out ./$package
    local exit_code=$?
    
    if [ $exit_code -eq 0 ]; then
        print_success "$description passed"
        
        # Generate coverage report
        if [ -f "coverage_${package//\//_}.out" ]; then
            coverage=$(go tool cover -func=coverage_${package//\//_}.out | grep total | awk '{print $3}')
            print_status "Coverage for $package: $coverage"
        fi
    else
        print_error "$description failed"
        return $exit_code
    fi
    
    return 0
}

# Track overall test results
overall_result=0

echo ""
print_status "Starting test execution..."
echo ""

# 1. Run Model Tests
print_status "📊 Testing Models..."
run_tests "models" "Model Tests"
if [ $? -ne 0 ]; then overall_result=1; fi

echo ""

# 2. Run Utils Tests
print_status "🔧 Testing Utils..."
run_tests "utils" "Utils Tests"
if [ $? -ne 0 ]; then overall_result=1; fi

echo ""

# 3. Run Middleware Tests
print_status "🛡️ Testing Middleware..."
run_tests "middleware" "Middleware Tests"
if [ $? -ne 0 ]; then overall_result=1; fi

echo ""

# 4. Run Controller Tests
print_status "🎮 Testing Controllers..."
run_tests "controllers" "Controller Tests"
if [ $? -ne 0 ]; then overall_result=1; fi

echo ""

# 5. Run Integration Tests
print_status "🔗 Testing Integration..."
go test -v -race -coverprofile=coverage_integration.out ./integration_test.go
if [ $? -eq 0 ]; then
    print_success "Integration Tests passed"
else
    print_error "Integration Tests failed"
    overall_result=1
fi

echo ""

# 6. Generate Combined Coverage Report
print_status "📈 Generating Coverage Report..."
echo "mode: set" > coverage_combined.out
for file in coverage_*.out; do
    if [ -f "$file" ] && [ "$file" != "coverage_combined.out" ]; then
        tail -n +2 "$file" >> coverage_combined.out
    fi
done

if [ -f "coverage_combined.out" ]; then
    total_coverage=$(go tool cover -func=coverage_combined.out | grep total | awk '{print $3}')
    print_status "Total Coverage: $total_coverage"
    
    # Generate HTML coverage report
    go tool cover -html=coverage_combined.out -o coverage.html
    print_status "HTML coverage report generated: coverage.html"
fi

echo ""

# 7. Run Benchmark Tests (if any exist)
print_status "⚡ Running Benchmarks..."
go test -bench=. -benchmem ./... > benchmark_results.txt 2>&1
if [ $? -eq 0 ]; then
    print_success "Benchmarks completed"
    print_status "Benchmark results saved to: benchmark_results.txt"
else
    print_warning "No benchmarks found or benchmarks failed"
fi

echo ""

# 8. Run Race Condition Detection
print_status "🏃 Running Race Condition Detection..."
go test -race ./...
if [ $? -eq 0 ]; then
    print_success "No race conditions detected"
else
    print_warning "Potential race conditions detected"
fi

echo ""

# 9. Run Static Analysis (if tools are available)
print_status "🔍 Running Static Analysis..."

# Check if golint is available
if command -v golint &> /dev/null; then
    print_status "Running golint..."
    golint ./... > lint_results.txt
    if [ -s lint_results.txt ]; then
        print_warning "Linting issues found. Check lint_results.txt"
    else
        print_success "No linting issues found"
    fi
else
    print_warning "golint not installed. Skipping linting."
fi

# Check if go vet is available (should be built-in)
print_status "Running go vet..."
go vet ./...
if [ $? -eq 0 ]; then
    print_success "go vet passed"
else
    print_warning "go vet found issues"
fi

echo ""

# 10. Test Build
print_status "🔨 Testing Build..."
go build -o test_build .
if [ $? -eq 0 ]; then
    print_success "Build successful"
    rm -f test_build
else
    print_error "Build failed"
    overall_result=1
fi

echo ""
echo "=============================================="

# Final Results
if [ $overall_result -eq 0 ]; then
    print_success "🎉 All tests passed successfully!"
    echo ""
    print_status "Summary:"
    print_status "- Model tests: ✅"
    print_status "- Utils tests: ✅"
    print_status "- Middleware tests: ✅"
    print_status "- Controller tests: ✅"
    print_status "- Integration tests: ✅"
    print_status "- Build test: ✅"
    
    if [ -f "coverage_combined.out" ]; then
        print_status "- Total coverage: $total_coverage"
    fi
else
    print_error "❌ Some tests failed. Please check the output above."
    exit 1
fi

echo ""
print_status "Test artifacts generated:"
print_status "- coverage.html (HTML coverage report)"
print_status "- coverage_*.out (Coverage data files)"
print_status "- benchmark_results.txt (Benchmark results)"
print_status "- lint_results.txt (Linting results)"

echo ""
print_status "To view coverage report: open coverage.html in your browser"
print_status "To clean up test artifacts: rm coverage*.out *.txt"
