package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	var (
		phone    = flag.String("phone", "13800138000", "用户手机号")
		password = flag.String("password", "admin123", "新密码")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := model.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 查找用户
	var user model.User
	if err := db.Where("phone = ?", *phone).First(&user).Error; err != nil {
		log.Fatalf("用户不存在: %v", err)
	}

	// 生成密码哈希
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("密码加密失败: %v", err)
	}

	// 更新密码
	if err := db.Model(&user).Update("password", string(passwordHash)).Error; err != nil {
		log.Fatalf("更新密码失败: %v", err)
	}

	fmt.Printf("密码重置成功！\n")
	fmt.Printf("手机号: %s\n", user.Phone)
	fmt.Printf("新密码: %s\n", *password)
}
