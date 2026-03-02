package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/database"
	infraRepo "github.com/Snowitty-Re/CNtunyuan/internal/infrastructure/repository"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	var (
		phone    = flag.String("phone", "13800138000", "user phone")
		password = flag.String("password", "admin123", "new password")
	)
	flag.Parse()

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Load config failed: %v", err)
	}

	db, err := database.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Connect database failed: %v", err)
	}

	userRepo := infraRepo.NewUserRepository(db)

	user, err := userRepo.FindByPhone(nil, *phone)
	if err != nil {
		log.Fatalf("User not found: %v", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Hash password failed: %v", err)
	}

	if err := userRepo.UpdatePassword(nil, user.ID, string(passwordHash)); err != nil {
		log.Fatalf("Update password failed: %v", err)
	}

	fmt.Printf("Password reset success!\n")
	fmt.Printf("Phone: %s\n", user.Phone)
	fmt.Printf("New Password: %s\n", *password)
}
