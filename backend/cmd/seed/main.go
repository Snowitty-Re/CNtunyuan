package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/internal/domain/repository"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/database"
	infraRepo "github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/repository"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"gorm.io/gorm"
)

// Seeder data seeder
type Seeder struct {
	db       *gorm.DB
	userRepo repository.UserRepository
}

// NewSeeder create seeder
func NewSeeder(db *gorm.DB) *Seeder {
	return &Seeder{
		db:       db,
		userRepo: infraRepo.NewUserRepository(db),
	}
}

// SeedAll seed all data
func (s *Seeder) SeedAll(ctx context.Context) error {
	if err := s.SeedOrganizations(ctx); err != nil {
		return fmt.Errorf("seed organizations failed: %w", err)
	}
	if err := s.SeedUsers(ctx); err != nil {
		return fmt.Errorf("seed users failed: %w", err)
	}
	return nil
}

// SeedOrganizations seed organizations
func (s *Seeder) SeedOrganizations(ctx context.Context) error {
	logger.Info("Start seeding organizations")

	var count int64
	if err := s.db.Model(&entity.Organization{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		logger.Info("Organizations already exist, skip")
		return nil
	}

	rootOrg := &entity.Organization{
		BaseEntity: entity.BaseEntity{
			ID: "00000000-0000-0000-0000-000000000001",
		},
		Name:   "Root Organization",
		Code:   "ROOT",
		Type:   entity.OrgTypeRoot,
		Level:  1,
		Status: entity.OrgStatusActive,
	}

	if err := s.db.Create(rootOrg).Error; err != nil {
		return err
	}

	logger.Info("Organizations seeded")
	return nil
}

// SeedUsers seed users
func (s *Seeder) SeedUsers(ctx context.Context) error {
	logger.Info("Start seeding users")

	var count int64
	if err := s.db.Model(&entity.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		logger.Info("Users already exist, skip")
		return nil
	}

	rootOrgID := "00000000-0000-0000-0000-000000000001"

	superAdmin, err := entity.NewUser("Super Admin", "13800138000", rootOrgID, entity.RoleSuperAdmin)
	if err != nil {
		return err
	}
	superAdmin.Email = "admin@cntuanyuan.com"
	if err := superAdmin.SetPassword("admin123"); err != nil {
		return err
	}

	if err := s.db.Create(superAdmin).Error; err != nil {
		return err
	}

	logger.Info("Users seeded")
	return nil
}

// Clean clean all data
func (s *Seeder) Clean(ctx context.Context) error {
	logger.Warn("Start cleaning data")

	tables := []string{
		"ty_user_permissions",
		"ty_users",
		"ty_organizations",
	}

	for _, table := range tables {
		if err := s.db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			return err
		}
	}

	logger.Info("Data cleaned")
	return nil
}

func main() {
	var (
		configPath = flag.String("config", "config/config.yaml", "config file path")
		all        = flag.Bool("all", false, "seed all data")
		orgs       = flag.Bool("orgs", false, "seed organizations only")
		users      = flag.Bool("users", false, "seed users only")
		clean      = flag.Bool("clean", false, "clean all data (dangerous!)")
	)
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Load config failed: %v", err)
	}

	logger.InitWithConfig(logger.Config{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	})

	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		logger.Fatal("Connect database failed", logger.Err(err))
	}

	seeder := NewSeeder(db)
	ctx := context.Background()

	if *clean {
		if err := seeder.Clean(ctx); err != nil {
			logger.Fatal("Clean data failed", logger.Err(err))
		}
		return
	}

	if *all {
		if err := seeder.SeedAll(ctx); err != nil {
			logger.Fatal("Seed data failed", logger.Err(err))
		}
	} else {
		if *orgs {
			if err := seeder.SeedOrganizations(ctx); err != nil {
				logger.Fatal("Seed organizations failed", logger.Err(err))
			}
		}
		if *users {
			if err := seeder.SeedUsers(ctx); err != nil {
				logger.Fatal("Seed users failed", logger.Err(err))
			}
		}
	}

	logger.Info("Seeding completed")
}
