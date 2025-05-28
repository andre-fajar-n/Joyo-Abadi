#!/bin/bash

# 🔄 Joyo Abadi Migration Script
# This script helps you manage database migrations for the multi-unit system

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ACTION="status"
DB_TYPE="sqlite"
DB_FILE="joyo_abadi.db"
VERBOSE=false
DRY_RUN=false
FORCE=false

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to show usage
show_usage() {
    echo "🔄 Joyo Abadi Migration Tool"
    echo ""
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Actions:"
    echo "  status    Check migration status (default)"
    echo "  up        Run migration (upgrade to multi-unit system)"
    echo "  down      Rollback migration (remove multi-unit system)"
    echo "  reset     Reset database completely"
    echo ""
    echo "Options:"
    echo "  -a, --action ACTION     Migration action (status|up|down|reset)"
    echo "  -d, --db TYPE          Database type (sqlite|postgres)"
    echo "  -f, --file FILE        SQLite database file (default: joyo_abadi.db)"
    echo "  -u, --url URL          PostgreSQL connection URL"
    echo "  -v, --verbose          Verbose output"
    echo "  -n, --dry-run          Show what would be done without executing"
    echo "  --force                Force dangerous operations"
    echo "  -h, --help             Show this help"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Check status"
    echo "  $0 -a up -v                         # Run migration with verbose output"
    echo "  $0 -a up -n                         # Dry run migration"
    echo "  $0 -a down --force                  # Rollback migration"
    echo "  $0 -d postgres -u 'host=...'        # Use PostgreSQL"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -a|--action)
            ACTION="$2"
            shift 2
            ;;
        -d|--db)
            DB_TYPE="$2"
            shift 2
            ;;
        -f|--file)
            DB_FILE="$2"
            shift 2
            ;;
        -u|--url)
            DB_URL="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -n|--dry-run)
            DRY_RUN=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validate action
case $ACTION in
    status|up|down|reset)
        ;;
    *)
        print_error "Invalid action: $ACTION"
        show_usage
        exit 1
        ;;
esac

# Build command
CMD="go run cmd/migrate/main.go"
CMD="$CMD -action=$ACTION"
CMD="$CMD -db=$DB_TYPE"

if [[ $DB_TYPE == "sqlite" ]]; then
    CMD="$CMD -file=$DB_FILE"
elif [[ $DB_TYPE == "postgres" && -n "$DB_URL" ]]; then
    CMD="$CMD -url=$DB_URL"
fi

if [[ $VERBOSE == true ]]; then
    CMD="$CMD -verbose"
fi

if [[ $DRY_RUN == true ]]; then
    CMD="$CMD -dry-run"
fi

if [[ $FORCE == true ]]; then
    CMD="$CMD -force"
fi

# Show what we're about to do
print_info "Joyo Abadi Migration Tool"
print_info "Action: $ACTION"
print_info "Database: $DB_TYPE"

if [[ $DB_TYPE == "sqlite" ]]; then
    print_info "Database file: $DB_FILE"
fi

if [[ $DRY_RUN == true ]]; then
    print_warning "DRY RUN MODE - No changes will be made"
fi

# Safety checks
case $ACTION in
    down)
        if [[ $FORCE != true && $DRY_RUN != true ]]; then
            print_warning "Rollback will remove multi-unit system and may cause data loss!"
            print_warning "Use --force to confirm or --dry-run to see what would happen"
            exit 1
        fi
        ;;
    reset)
        if [[ $FORCE != true && $DRY_RUN != true ]]; then
            print_error "Reset will DELETE ALL DATA in the database!"
            print_error "Use --force to confirm or --dry-run to see what would happen"
            exit 1
        fi
        ;;
esac

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check if migration file exists
if [[ ! -f "cmd/migrate/main.go" ]]; then
    print_error "Migration file not found: cmd/migrate/main.go"
    print_info "Make sure you're running this from the project root directory"
    exit 1
fi

# For SQLite, check if database file exists (except for reset)
if [[ $DB_TYPE == "sqlite" && $ACTION != "reset" && ! -f "$DB_FILE" ]]; then
    print_warning "Database file not found: $DB_FILE"
    if [[ $ACTION == "up" ]]; then
        print_info "A new database will be created"
    else
        print_error "Cannot perform $ACTION on non-existent database"
        exit 1
    fi
fi

# Backup suggestion for production
if [[ $ACTION == "up" && $FORCE != true && $DRY_RUN != true ]]; then
    print_warning "IMPORTANT: Backup your database before running migration!"
    
    if [[ $DB_TYPE == "sqlite" ]]; then
        print_info "Backup command: cp $DB_FILE ${DB_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
    else
        print_info "Backup command: pg_dump joyo_abadi > backup_$(date +%Y%m%d_%H%M%S).sql"
    fi
    
    echo ""
    read -p "Have you backed up your database? (y/N): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Please backup your database first, then run the migration again"
        exit 0
    fi
fi

# Run the migration
print_info "Running migration command..."
if [[ $VERBOSE == true ]]; then
    print_info "Command: $CMD"
fi

echo ""
eval $CMD
RESULT=$?

echo ""
if [[ $RESULT -eq 0 ]]; then
    case $ACTION in
        up)
            if [[ $DRY_RUN != true ]]; then
                print_success "Migration completed successfully!"
                print_info "Your database now supports multi-unit products!"
                print_info "Next steps:"
                print_info "  1. Start the application: go run main.go"
                print_info "  2. Create a product and add multiple units"
                print_info "  3. Read the documentation: MULTI_UNIT_SYSTEM.md"
            fi
            ;;
        down)
            if [[ $DRY_RUN != true ]]; then
                print_success "Migration rolled back successfully!"
                print_warning "Multi-unit system has been removed"
            fi
            ;;
        reset)
            if [[ $DRY_RUN != true ]]; then
                print_success "Database reset completed!"
                print_info "Fresh database with multi-unit system ready!"
            fi
            ;;
        status)
            print_success "Status check completed!"
            ;;
    esac
else
    print_error "Migration failed with exit code $RESULT"
    print_info "Check the error messages above for details"
    exit $RESULT
fi
