// Seeder reads scripts/seed/seed_data.json and upserts dictionary entries
// into PostgreSQL. Also upserts a default admin user. Safe to run multiple
// times (idempotent).
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	dictq "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/infrastructure/postgres/sqlc"
	identityq "github.com/iqbaleff214/kamus-banjar-api-2/internal/identity/infrastructure/postgres/sqlc"
)

// ─── JSON seed schema ─────────────────────────────────────────────────────────

type seedMeta struct {
	Source    string `json:"source"`
	Edition   string `json:"edition"`
	Publisher string `json:"publisher"`
	Year      int    `json:"year"`
	ISBN      string `json:"isbn"`
	Dialect   string `json:"dialect"`
}

type seedExample struct {
	Banjar     string `json:"banjar"`
	Indonesian string `json:"indonesian"`
}

type seedEntry struct {
	Banjar            string        `json:"banjar"`
	BanjarSyllabified string        `json:"banjar_syllabified"`
	WordClass         string        `json:"word_class"`
	WordClassFull     string        `json:"word_class_full"`
	Definitions       []string      `json:"definitions"`
	Examples          []seedExample `json:"examples"`
	IsDerived         bool          `json:"is_derived"`
	HomonymNumber     int           `json:"homonym_number"`
	Dialect           string        `json:"dialect"`
	Source            string        `json:"source"`
	SourceReference   string        `json:"source_reference"`
	DerivedForms      []seedEntry   `json:"derived_forms"`
}

type seedFile struct {
	Meta    seedMeta    `json:"meta"`
	Entries []seedEntry `json:"entries"`
}

// ─── main ─────────────────────────────────────────────────────────────────────

func main() {
	_ = godotenv.Load()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		mustEnv("DB_USER"), mustEnv("DB_PASS"),
		mustEnv("DB_HOST"), mustEnv("DB_PORT"), mustEnv("DB_NAME"),
	)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	db := stdlib.OpenDBFromPool(pool)

	if err := seedAdmin(ctx, identityq.New(db)); err != nil {
		log.Fatalf("seed admin: %v", err)
	}

	data, err := loadSeedFile("scripts/seed/seed_data.json")
	if err != nil {
		log.Fatalf("load seed: %v", err)
	}

	queries := dictq.New(db)
	log.Printf("Seeding %d entries...", len(data.Entries))
	inserted, updated := 0, 0

	for _, entry := range data.Entries {
		rootID, wasNew, err := upsertEntry(ctx, queries, entry, nil)
		if err != nil {
			log.Printf("WARN: skip %q: %v", entry.Banjar, err)
			continue
		}
		if wasNew {
			inserted++
		} else {
			updated++
		}

		for _, derived := range entry.DerivedForms {
			_, wasNew2, err := upsertEntry(ctx, queries, derived, &rootID)
			if err != nil {
				log.Printf("WARN: skip derived %q: %v", derived.Banjar, err)
				continue
			}
			if wasNew2 {
				inserted++
			} else {
				updated++
			}
		}
	}

	log.Printf("Done. inserted=%d updated=%d", inserted, updated)
}

// ─── admin user ───────────────────────────────────────────────────────────────

func seedAdmin(ctx context.Context, queries *identityq.Queries) error {
	email := getEnvOrDefault("ADMIN_EMAIL", "iqbaleff214@gmail.com")
	password := getEnvOrDefault("ADMIN_PASSWORD", "Admin1234!")
	name := getEnvOrDefault("ADMIN_NAME", "Admin")

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	_, err = queries.CreateUser(ctx, identityq.CreateUserParams{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         identityq.UserRoleAdmin,
		IsActive:     true,
		EmailVerifiedAt: sql.NullTime{
			Time:  now,
			Valid: true,
		},
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		// Ignore duplicate — admin already exists
		if isUniqueViolation(err) {
			log.Printf("admin user %q already exists, skipping", email)
			return nil
		}
		return fmt.Errorf("create admin: %w", err)
	}

	log.Printf("admin user created: email=%s password=%s", email, password)
	return nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "23505") || contains(err.Error(), "unique"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ─── upsert logic ─────────────────────────────────────────────────────────────

func upsertEntry(ctx context.Context, queries *dictq.Queries, entry seedEntry, rootID *uuid.UUID) (uuid.UUID, bool, error) {
	wordClass, err := normalizeWordClass(entry.WordClass)
	if err != nil {
		return uuid.UUID{}, false, fmt.Errorf("unknown word class %q: %w", entry.WordClass, err)
	}

	dialect := dictq.Dialect("hulu")
	homonym := int16(entry.HomonymNumber)
	if homonym < 1 {
		homonym = 1
	}
	isRoot := !entry.IsDerived
	if rootID == nil {
		isRoot = true
	}

	var rootWordID uuid.NullUUID
	if rootID != nil {
		rootWordID = uuid.NullUUID{UUID: *rootID, Valid: true}
	}

	syllabified := sql.NullString{}
	if entry.BanjarSyllabified != "" && entry.BanjarSyllabified != entry.Banjar {
		syllabified = sql.NullString{String: entry.BanjarSyllabified, Valid: true}
	}

	srcRef := sql.NullString{}
	if entry.SourceReference != "" {
		srcRef = sql.NullString{String: entry.SourceReference, Valid: true}
	}

	now := time.Now().UTC()
	row, err := queries.UpsertWord(ctx, dictq.UpsertWordParams{
		ID:                uuid.New(),
		Banjar:            entry.Banjar,
		BanjarSyllabified: syllabified,
		Dialect:           dialect,
		WordClass:         wordClass,
		HomonymNumber:     homonym,
		IsRoot:            isRoot,
		RootWordID:        rootWordID,
		Status:            dictq.WordStatusActive,
		Source:            dictq.WordSourceSeeded,
		SourceReference:   srcRef,
		CreatedBy:         uuid.NullUUID{},
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		return uuid.UUID{}, false, err
	}

	wasNew := row.CreatedAt.Equal(now) || row.UpdatedAt.Equal(now)

	_ = queries.DeleteDefinitionsByWordID(ctx, row.ID)
	for i, meaning := range entry.Definitions {
		if meaning == "" {
			continue
		}
		_, _ = queries.UpsertDefinition(ctx, dictq.UpsertDefinitionParams{
			ID:        uuid.New(),
			WordID:    row.ID,
			Meaning:   meaning,
			SortOrder: int16(i + 1),
			Source:    dictq.WordSourceSeeded,
			Upvotes:   0,
			Downvotes: 0,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	_ = queries.DeleteExamplesByWordID(ctx, row.ID)
	for _, ex := range entry.Examples {
		if ex.Banjar == "" {
			continue
		}
		_, _ = queries.UpsertExample(ctx, dictq.UpsertExampleParams{
			ID:                    uuid.New(),
			WordID:                row.ID,
			BanjarSentence:        ex.Banjar,
			IndonesianTranslation: ex.Indonesian,
			Source:                dictq.WordSourceSeeded,
			CreatedAt:             now,
			UpdatedAt:             now,
		})
	}

	return row.ID, wasNew, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func normalizeWordClass(s string) (dictq.WordClass, error) {
	switch s {
	case "n":
		return dictq.WordClassN, nil
	case "v":
		return dictq.WordClassV, nil
	case "a":
		return dictq.WordClassA, nil
	case "adv":
		return dictq.WordClassAdv, nil
	case "p":
		return dictq.WordClassP, nil
	case "pb":
		return dictq.WordClassPb, nil
	case "ki":
		return dictq.WordClassKi, nil
	case "num":
		return dictq.WordClassNum, nil
	case "pron":
		return dictq.WordClassPron, nil
	default:
		// Strip OCR artifacts (e.g. "n`") and retry
		cleaned := strings.TrimRight(s, "`'\"")
		if cleaned != s {
			return normalizeWordClass(cleaned)
		}
		return dictq.WordClassN, fmt.Errorf("unknown")
	}
}

func loadSeedFile(path string) (*seedFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var data seedFile
	return &data, json.NewDecoder(f).Decode(&data)
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing env var: %s", key)
	}
	return v
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
