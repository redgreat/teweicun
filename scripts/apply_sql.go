package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/redgreat/teweicun/internal/config"
	"github.com/redgreat/teweicun/pkg/database"
	"github.com/redgreat/teweicun/pkg/logger"
)

func main() {
	var cfgPath string
	var sqlFile string
	flag.StringVar(&cfgPath, "c", "", "path to config file (default: configs/config.yaml)")
	flag.StringVar(&sqlFile, "f", "", "path to .sql file")
	flag.Parse()

	if strings.TrimSpace(sqlFile) == "" {
		_, _ = fmt.Fprintln(os.Stderr, "missing -f <sql file>")
		os.Exit(2)
	}

	if err := config.LoadConfig(cfgPath); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "load config failed:", err)
		os.Exit(1)
	}

	_ = logger.InitLogger("info", true, "")
	defer logger.Sync()

	if err := database.InitPostgres(config.GlobalConfig.Database); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "init database failed:", err)
		os.Exit(1)
	}
	defer database.ClosePostgres()

	stmts, err := readSQLStatements(sqlFile)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "read sql failed:", err)
		os.Exit(1)
	}

	migrationID, err := readMigrationID(sqlFile)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "read migration id failed:", err)
		os.Exit(1)
	}
	if migrationID == "" {
		migrationID = filepath.Base(sqlFile)
	}
	checksum, err := fileChecksumSHA256(sqlFile)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "calc checksum failed:", err)
		os.Exit(1)
	}

	ctx := context.Background()
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "begin tx failed:", err)
		os.Exit(1)
	}
	defer tx.Rollback(ctx)

	for _, s := range stmts {
		if _, err := tx.Exec(ctx, s); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "exec failed:", err)
			os.Exit(1)
		}
	}

	if _, err := tx.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migration_history (
			migration_id varchar(255) PRIMARY KEY,
			filename varchar(255) NOT NULL,
			file_checksum varchar(128) NOT NULL,
			applied_at timestamp with time zone NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "ensure schema_migration_history failed:", err)
		os.Exit(1)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO schema_migration_history (migration_id, filename, file_checksum)
		VALUES ($1, $2, $3)
		ON CONFLICT (migration_id) DO NOTHING
	`, migrationID, filepath.Base(sqlFile), checksum); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "record migration history failed:", err)
		os.Exit(1)
	}

	if err := tx.Commit(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "commit failed:", err)
		os.Exit(1)
	}

	fmt.Println("ok")
}

func readSQLStatements(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var b strings.Builder
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "--") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	raw := b.String()
	// Split by ; but respect dollar-quoted strings (e.g. $procedure$...$procedure$)
	parts := splitSQLStatements(raw)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		up := strings.ToUpper(strings.TrimSpace(s))
		if up == "BEGIN" || up == "COMMIT" {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

// splitSQLStatements splits SQL text by ; but respects dollar-quoted blocks like $tag$...$tag$
func splitSQLStatements(raw string) []string {
	var result []string
	var current strings.Builder
	inDollar := false
	dollarTag := ""

	i := 0
	for i < len(raw) {
		ch := raw[i]

		if !inDollar && ch == '$' {
			// Look for matching $tag$ pattern
			rest := raw[i:]
			if endDollar := findDollarEnd(rest); endDollar > 0 && endDollar < len(rest) && rest[endDollar] == '$' {
				inDollar = true
				dollarTag = rest[:endDollar+1]
				current.WriteByte(ch)
				i++
				continue
			}
		} else if inDollar && ch == '$' {
			// Check if this matches the closing $tag$
			rest := raw[i:]
			if strings.HasPrefix(rest, dollarTag) {
				inDollar = false
				// Write the closing dollar tag
				for k := 0; k < len(dollarTag); k++ {
					current.WriteByte(raw[i])
					i++
				}
				continue
			}
		}

		if !inDollar && ch == ';' {
			result = append(result, current.String())
			current.Reset()
			i++
			continue
		}

		current.WriteByte(ch)
		i++
	}

	// Last fragment
	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// findDollarEnd finds the closing $ in a $tag$ pattern starting from pos 0
// returns the index of the closing $ (exclusive in the original slice)
func findDollarEnd(s string) int {
	if len(s) < 2 || s[0] != '$' {
		return -1
	}
	// Look for next $ that's not at position 0
	for j := 1; j < len(s); j++ {
		if s[j] == '$' {
			return j
		}
	}
	return -1
}

func readMigrationID(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "--") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "--"))
		if strings.HasPrefix(line, "MIGRATION_ID:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "MIGRATION_ID:")), nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return "", nil
}

func fileChecksumSHA256(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
