// Seeder reads scripts/seed/seed_data.json and upserts dictionary entries
// into PostgreSQL. Safe to run multiple times (idempotent).
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	q "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/infrastructure/postgres/sqlc"
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
	Banjar    string `json:"banjar"`
	Indonesian string `json:"indonesian"`
}

type seedEntry struct {
	Banjar           string        `json:"banjar"`
	BanjarSyllabified string       `json:"banjar_syllabified"`
	WordClass        string        `json:"word_class"`
	WordClassFull    string        `json:"word_class_full"`
	Definitions      []string      `json:"definitions"`
	Examples         []seedExample `json:"examples"`
	IsDerived        bool          `json:"is_derived"`
	HomonymNumber    int           `json:"homonym_number"`
	Dialect          string        `json:"dialect"`
	Source           string        `json:"source"`
	SourceReference  string        `json:"source_reference"`
	DerivedForms     []seedEntry   `json:"derived_forms"`
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
	queries := q.New(db)

	data, err := loadSeedFile("scripts/seed/seed_data.json")
	if err != nil {
		log.Fatalf("load seed: %v", err)
	}

	log.Printf("Seeding %d root entries...", len(data.Entries))
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

		// Derived forms
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

// ─── upsert logic ─────────────────────────────────────────────────────────────

func upsertEntry(ctx context.Context, queries *q.Queries, entry seedEntry, rootID *uuid.UUID) (uuid.UUID, bool, error) {
	wordClass, err := normalizeWordClass(entry.WordClass)
	if err != nil {
		return uuid.UUID{}, false, fmt.Errorf("unknown word class %q: %w", entry.WordClass, err)
	}

	dialect := q.Dialect("hulu")
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
	row, err := queries.UpsertWord(ctx, q.UpsertWordParams{
		ID:                uuid.New(),
		Banjar:            entry.Banjar,
		BanjarSyllabified: syllabified,
		Dialect:           dialect,
		WordClass:         wordClass,
		HomonymNumber:     homonym,
		IsRoot:            isRoot,
		RootWordID:        rootWordID,
		Status:            q.WordStatusActive,
		Source:            q.WordSourceSeeded,
		SourceReference:   srcRef,
		CreatedBy:         uuid.NullUUID{},
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		return uuid.UUID{}, false, err
	}

	wasNew := row.CreatedAt.Equal(now) || row.UpdatedAt.Equal(now)

	// Replace definitions
	_ = queries.DeleteDefinitionsByWordID(ctx, row.ID)
	for i, meaning := range entry.Definitions {
		if meaning == "" {
			continue
		}
		_, _ = queries.UpsertDefinition(ctx, q.UpsertDefinitionParams{
			ID:        uuid.New(),
			WordID:    row.ID,
			Meaning:   meaning,
			SortOrder: int16(i + 1),
			Source:    q.WordSourceSeeded,
			Upvotes:   0,
			Downvotes: 0,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	// Replace examples
	_ = queries.DeleteExamplesByWordID(ctx, row.ID)
	for _, ex := range entry.Examples {
		if ex.Banjar == "" {
			continue
		}
		_, _ = queries.UpsertExample(ctx, q.UpsertExampleParams{
			ID:                    uuid.New(),
			WordID:                row.ID,
			BanjarSentence:        ex.Banjar,
			IndonesianTranslation: ex.Indonesian,
			Source:                q.WordSourceSeeded,
			CreatedAt:             now,
			UpdatedAt:             now,
		})
	}

	return row.ID, wasNew, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func normalizeWordClass(s string) (q.WordClass, error) {
	switch s {
	case "n":
		return q.WordClassN, nil
	case "v":
		return q.WordClassV, nil
	case "a":
		return q.WordClassA, nil
	case "adv":
		return q.WordClassAdv, nil
	case "p":
		return q.WordClassP, nil
	case "pb":
		return q.WordClassPb, nil
	case "ki":
		return q.WordClassKi, nil
	default:
		// Map numeralia/pronomina to closest equivalent for seed data
		return q.WordClassN, fmt.Errorf("unknown")
	}
}

func loadSeedFile(path string) (*seedFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
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
