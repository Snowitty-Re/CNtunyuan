package database

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"gorm.io/gorm"
)

func RunBootstrapMigration(db *gorm.DB, dbType config.DatabaseType) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	sqlPath, err := resolveBootstrapMigrationPath(dbType)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		return fmt.Errorf("读取初始化 SQL 失败: %w", err)
	}

	statements := splitSQLStatements(string(content))
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		if err := db.Exec(trimmed).Error; err != nil {
			return fmt.Errorf("执行初始化 SQL 失败: %w", err)
		}
	}
	return nil
}

func resolveBootstrapMigrationPath(dbType config.DatabaseType) (string, error) {
	name := ""
	switch dbType {
	case config.DatabaseTypePostgres:
		name = filepath.Join("migrations", "postgres", "00_bootstrap.sql")
	case config.DatabaseTypeMySQL:
		name = filepath.Join("migrations", "mysql", "00_bootstrap.sql")
	default:
		return "", fmt.Errorf("unsupported database type: %s", dbType)
	}

	candidates := []string{
		filepath.Join(".", name),
		filepath.Join("backend", name),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("未找到初始化 SQL 文件: %s", name)
}

func splitSQLStatements(content string) []string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024*10)

	var (
		builder       strings.Builder
		statements    []string
		inSingleQuote bool
		inDoubleQuote bool
	)

	flush := func() {
		stmt := strings.TrimSpace(builder.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
		builder.Reset()
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		for i := 0; i < len(line); i++ {
			ch := line[i]
			prev := byte(0)
			if i > 0 {
				prev = line[i-1]
			}

			switch ch {
			case '\'':
				if !inDoubleQuote && prev != '\\' {
					inSingleQuote = !inSingleQuote
				}
			case '"':
				if !inSingleQuote && prev != '\\' {
					inDoubleQuote = !inDoubleQuote
				}
			case ';':
				if !inSingleQuote && !inDoubleQuote {
					builder.WriteByte(ch)
					flush()
					continue
				}
			}
			builder.WriteByte(ch)
		}
		builder.WriteByte('\n')
	}

	if builder.Len() > 0 {
		flush()
	}

	return statements
}
