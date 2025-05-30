package main

import (
	"flag"
	"fmt"
	"joyo-abadi/migrations"
	"joyo-abadi/models"
	"joyo-abadi/utils"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Command line flags
	var (
		action   = flag.String("action", "up", "Migration action: up, down, status")
		dbType   = flag.String("db", "sqlite", "Database type: sqlite, postgres")
		dbURL    = flag.String("url", "", "Database URL (for postgres)")
		dbFile   = flag.String("file", "joyo_abadi.db", "Database file (for sqlite)")
		force    = flag.Bool("force", false, "Force migration (use with caution)")
		dryRun   = flag.Bool("dry-run", false, "Show what would be done without executing")
		verbose  = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	// Initialize logger
	utils.InitLogger()

	if *verbose {
		log.Println("🚀 Joyo Abadi Migration Tool")
		log.Printf("Action: %s", *action)
		log.Printf("Database: %s", *dbType)
		log.Printf("Dry Run: %t", *dryRun)
	}

	// Connect to database
	db, err := connectDatabase(*dbType, *dbURL, *dbFile)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Execute migration action
	switch *action {
	case "up":
		runMigrationUp(db, *dryRun, *verbose)
	case "down":
		runMigrationDown(db, *force, *dryRun, *verbose)
	case "status":
		showMigrationStatus(db, *verbose)
	case "reset":
		resetDatabase(db, *force, *dryRun, *verbose)
	default:
		log.Fatalf("❌ Unknown action: %s. Use: up, down, status, reset", *action)
	}
}

func connectDatabase(dbType, dbURL, dbFile string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch dbType {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{
			Logger: &utils.GormLogger{},
		})
	case "postgres":
		if dbURL == "" {
			// Build URL from environment variables
			dbURL = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
				getEnv("DB_HOST", "localhost"),
				getEnv("DB_USER", "postgres"),
				getEnv("DB_PASSWORD", ""),
				getEnv("DB_NAME", "joyo_abadi"),
				getEnv("DB_PORT", "5432"),
				getEnv("DB_SSLMODE", "disable"),
			)
		}
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{
			Logger: &utils.GormLogger{},
		})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	return db, err
}

func runMigrationUp(db *gorm.DB, dryRun, verbose bool) {
	if verbose {
		log.Println("📈 Running migration UP...")
	}

	if dryRun {
		log.Println("🔍 DRY RUN - Would perform the following actions:")
		log.Println("  1. Create product_units table")
		log.Println("  2. Add new columns to products table (base_unit_name, is_active)")
		log.Println("  3. Add new columns to transactions table (product_unit_id, unit_name, unit_price, base_quantity, notes)")
		log.Println("  4. Migrate existing products to have base units")
		log.Println("  5. Update existing transactions with unit information")
		return
	}

	// Run the migration
	if err := migrations.MigrateProductUnits(db); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Migration completed successfully!")
	log.Println("🎉 Your database now supports multi-unit products!")
}

func runMigrationDown(db *gorm.DB, force, dryRun, verbose bool) {
	if verbose {
		log.Println("📉 Running migration DOWN...")
	}

	if !force && !dryRun {
		log.Println("⚠️  WARNING: This will remove the multi-unit system and may cause data loss!")
		log.Println("Use --force flag to confirm, or --dry-run to see what would be done.")
		return
	}

	if dryRun {
		log.Println("🔍 DRY RUN - Would perform the following actions:")
		log.Println("  1. Drop product_units table")
		log.Println("  2. Remove new columns from products and transactions tables (manual cleanup required)")
		log.Println("⚠️  Note: Some data may be lost in this process")
		return
	}

	// Run the rollback
	if err := migrations.RollbackProductUnits(db); err != nil {
		log.Fatalf("❌ Rollback failed: %v", err)
	}

	log.Println("✅ Migration rolled back successfully!")
	log.Println("⚠️  Note: Manual cleanup of new columns may be required")
}

func showMigrationStatus(db *gorm.DB, verbose bool) {
	if verbose {
		log.Println("📊 Checking migration status...")
	}

	// Check if product_units table exists
	hasProductUnits := db.Migrator().HasTable(&models.ProductUnit{})
	
	// Check if new columns exist
	hasBaseUnitName := db.Migrator().HasColumn(&models.Product{}, "base_unit_name")
	hasIsActive := db.Migrator().HasColumn(&models.Product{}, "is_active")
	hasProductUnitID := db.Migrator().HasColumn(&models.Transaction{}, "product_unit_id")

	fmt.Println("📋 Migration Status:")
	fmt.Printf("  ProductUnits table: %s\n", statusIcon(hasProductUnits))
	fmt.Printf("  Product.BaseUnitName: %s\n", statusIcon(hasBaseUnitName))
	fmt.Printf("  Product.IsActive: %s\n", statusIcon(hasIsActive))
	fmt.Printf("  Transaction.ProductUnitID: %s\n", statusIcon(hasProductUnitID))

	if hasProductUnits && hasBaseUnitName && hasIsActive && hasProductUnitID {
		fmt.Println("✅ Multi-unit system is fully migrated")
		
		// Show some statistics
		var productCount, unitCount int64
		db.Model(&models.Product{}).Count(&productCount)
		db.Model(&models.ProductUnit{}).Count(&unitCount)
		
		fmt.Printf("📊 Statistics:\n")
		fmt.Printf("  Products: %d\n", productCount)
		fmt.Printf("  Product Units: %d\n", unitCount)
		
		if productCount > 0 {
			fmt.Printf("  Average units per product: %.1f\n", float64(unitCount)/float64(productCount))
		}
	} else {
		fmt.Println("❌ Multi-unit system is not fully migrated")
		fmt.Println("💡 Run: go run cmd/migrate/main.go -action=up")
	}
}

func resetDatabase(db *gorm.DB, force, dryRun, verbose bool) {
	if verbose {
		log.Println("🔄 Resetting database...")
	}

	if !force && !dryRun {
		log.Println("⚠️  WARNING: This will completely reset the database and delete ALL data!")
		log.Println("Use --force flag to confirm, or --dry-run to see what would be done.")
		return
	}

	if dryRun {
		log.Println("🔍 DRY RUN - Would perform the following actions:")
		log.Println("  1. Drop all tables")
		log.Println("  2. Recreate all tables with latest schema")
		log.Println("⚠️  ALL DATA WILL BE LOST!")
		return
	}

	// Drop all tables
	db.Migrator().DropTable(&models.Transaction{})
	db.Migrator().DropTable(&models.ProductUnit{})
	db.Migrator().DropTable(&models.Product{})
	db.Migrator().DropTable(&models.User{})

	// Recreate with latest schema
	err := db.AutoMigrate(&models.User{}, &models.Product{}, &models.ProductUnit{}, &models.Transaction{})
	if err != nil {
		log.Fatalf("❌ Reset failed: %v", err)
	}

	log.Println("✅ Database reset completed!")
	log.Println("🎉 Fresh database with multi-unit system ready!")
}

func statusIcon(exists bool) string {
	if exists {
		return "✅ Present"
	}
	return "❌ Missing"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
