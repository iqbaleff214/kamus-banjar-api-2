package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	q "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/infrastructure/postgres/sqlc"
)

type PostgresWordRepository struct {
	queries *q.Queries
}

func NewPostgresWordRepository(pool *pgxpool.Pool) *PostgresWordRepository {
	db := stdlib.OpenDBFromPool(pool)
	return &PostgresWordRepository{queries: q.New(db)}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (r *PostgresWordRepository) Create(ctx context.Context, word *domain.Word) error {
	row, err := r.queries.CreateWord(ctx, toCreateParams(word))
	if err != nil {
		if isUniqueViolation(err) {
			return domain.ErrWordConflict
		}
		return err
	}
	word.ID = row.ID
	// persist definitions and examples
	return r.upsertChildren(ctx, word)
}

// ─── FindByID ─────────────────────────────────────────────────────────────────

func (r *PostgresWordRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Word, error) {
	row, err := r.queries.GetWordByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrWordNotFound
		}
		return nil, err
	}
	return r.hydrate(ctx, row)
}

// ─── FindAll ──────────────────────────────────────────────────────────────────

func (r *PostgresWordRepository) FindAll(ctx context.Context, filter domain.WordFilter, page, perPage int) ([]*domain.Word, int, error) {
	params := listParams(filter, page, perPage)

	rows, err := r.queries.ListWords(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := r.queries.CountWords(ctx, q.CountWordsParams{
		Column1: params.Column1,
		Column2: params.Column2,
		Column3: params.Column3,
		Column4: params.Column4,
	})
	if err != nil {
		return nil, 0, err
	}
	words, err := r.hydrateMany(ctx, rows)
	return words, int(total), err
}

// ─── Search ───────────────────────────────────────────────────────────────────

func (r *PostgresWordRepository) Search(ctx context.Context, query string, filter domain.WordFilter, page, perPage int) ([]*domain.Word, int, error) {
	wc, src, stat, isRootStr := filterStrings(filter)
	limit := int32(perPage)
	offset := int32((page - 1) * perPage)

	rows, err := r.queries.SearchWords(ctx, q.SearchWordsParams{
		Column1: query,
		Column2: wc,
		Column3: isRootStr,
		Column4: src,
		Column5: stat,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := r.queries.CountSearchWords(ctx, q.CountSearchWordsParams{
		Column1: query,
		Column2: wc,
		Column3: isRootStr,
		Column4: src,
		Column5: stat,
	})
	if err != nil {
		return nil, 0, err
	}
	words, err := r.hydrateMany(ctx, rows)
	return words, int(total), err
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (r *PostgresWordRepository) Update(ctx context.Context, word *domain.Word) error {
	var rootWordID uuid.NullUUID
	if word.RootWordID != nil {
		rootWordID = uuid.NullUUID{UUID: *word.RootWordID, Valid: true}
	}
	var srcRef sql.NullString
	if word.SourceReference != "" {
		srcRef = sql.NullString{String: word.SourceReference, Valid: true}
	}
	_, err := r.queries.UpdateWord(ctx, q.UpdateWordParams{
		ID:                word.ID,
		Banjar:            word.Banjar,
		BanjarSyllabified: sql.NullString{String: word.BanjarSyllabified, Valid: word.BanjarSyllabified != ""},
		WordClass:         q.WordClass(word.WordClass),
		HomonymNumber:     int16(word.HomonymNumber),
		IsRoot:            word.IsRoot,
		RootWordID:        rootWordID,
		Status:            q.WordStatus(word.Status),
		Source:            q.WordSource(word.Source),
		SourceReference:   srcRef,
		UpdatedAt:         time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrWordNotFound
		}
		return err
	}
	// Replace definitions and examples
	if err := r.queries.DeleteDefinitionsByWordID(ctx, word.ID); err != nil {
		return err
	}
	if err := r.queries.DeleteExamplesByWordID(ctx, word.ID); err != nil {
		return err
	}
	return r.upsertChildren(ctx, word)
}

// ─── SoftDelete ───────────────────────────────────────────────────────────────

func (r *PostgresWordRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.queries.SoftDeleteWord(ctx, id)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (r *PostgresWordRepository) upsertChildren(ctx context.Context, word *domain.Word) error {
	now := time.Now().UTC()
	for _, def := range word.Definitions {
		if _, err := r.queries.UpsertDefinition(ctx, q.UpsertDefinitionParams{
			ID:        def.ID,
			WordID:    word.ID,
			Meaning:   def.Meaning,
			SortOrder: int16(def.SortOrder),
			Source:    q.WordSource(def.Source),
			Upvotes:   int32(def.Upvotes),
			Downvotes: int32(def.Downvotes),
			CreatedAt: now,
			UpdatedAt: now,
		}); err != nil {
			return err
		}
	}
	for _, ex := range word.Examples {
		if _, err := r.queries.UpsertExample(ctx, q.UpsertExampleParams{
			ID:                    ex.ID,
			WordID:                word.ID,
			BanjarSentence:        ex.BanjarSentence,
			IndonesianTranslation: ex.IndonesianTranslation,
			Source:                q.WordSource(ex.Source),
			CreatedAt:             now,
			UpdatedAt:             now,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresWordRepository) hydrate(ctx context.Context, row q.Word) (*domain.Word, error) {
	defs, err := r.queries.GetDefinitionsByWordID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	exs, err := r.queries.GetExamplesByWordID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return toDomain(row, defs, exs), nil
}

func (r *PostgresWordRepository) hydrateMany(ctx context.Context, rows []q.Word) ([]*domain.Word, error) {
	out := make([]*domain.Word, 0, len(rows))
	for _, row := range rows {
		w, err := r.hydrate(ctx, row)
		if err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, nil
}

func toDomain(row q.Word, defs []q.Definition, exs []q.Example) *domain.Word {
	w := &domain.Word{
		ID:            row.ID,
		Banjar:        row.Banjar,
		Dialect:       domain.Dialect(row.Dialect),
		WordClass:     domain.WordClass(row.WordClass),
		HomonymNumber: int(row.HomonymNumber),
		IsRoot:        row.IsRoot,
		Status:        domain.WordStatus(row.Status),
		Source:        domain.Source(row.Source),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
		Definitions:   make([]*domain.Definition, 0, len(defs)),
		Examples:      make([]*domain.Example, 0, len(exs)),
		RelatedWords:  make([]uuid.UUID, 0),
	}
	if row.BanjarSyllabified.Valid {
		w.BanjarSyllabified = row.BanjarSyllabified.String
	}
	if row.SourceReference.Valid {
		w.SourceReference = row.SourceReference.String
	}
	if row.RootWordID.Valid {
		id := row.RootWordID.UUID
		w.RootWordID = &id
	}
	if row.CreatedBy.Valid {
		id := row.CreatedBy.UUID
		w.CreatedBy = &id
	}
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		w.DeletedAt = &t
	}
	for _, d := range defs {
		w.Definitions = append(w.Definitions, &domain.Definition{
			ID:        d.ID,
			Meaning:   d.Meaning,
			SortOrder: int(d.SortOrder),
			Source:    domain.Source(d.Source),
			Upvotes:   int(d.Upvotes),
			Downvotes: int(d.Downvotes),
		})
	}
	for _, e := range exs {
		w.Examples = append(w.Examples, &domain.Example{
			ID:                    e.ID,
			BanjarSentence:        e.BanjarSentence,
			IndonesianTranslation: e.IndonesianTranslation,
			Source:                domain.Source(e.Source),
		})
	}
	return w
}

func toCreateParams(word *domain.Word) q.CreateWordParams {
	var rootWordID uuid.NullUUID
	if word.RootWordID != nil {
		rootWordID = uuid.NullUUID{UUID: *word.RootWordID, Valid: true}
	}
	var createdBy uuid.NullUUID
	if word.CreatedBy != nil {
		createdBy = uuid.NullUUID{UUID: *word.CreatedBy, Valid: true}
	}
	return q.CreateWordParams{
		ID:                word.ID,
		Banjar:            word.Banjar,
		BanjarSyllabified: sql.NullString{String: word.BanjarSyllabified, Valid: word.BanjarSyllabified != ""},
		Dialect:           q.Dialect(word.Dialect),
		WordClass:         q.WordClass(word.WordClass),
		HomonymNumber:     int16(word.HomonymNumber),
		IsRoot:            word.IsRoot,
		RootWordID:        rootWordID,
		Status:            q.WordStatus(word.Status),
		Source:            q.WordSource(word.Source),
		SourceReference:   sql.NullString{String: word.SourceReference, Valid: word.SourceReference != ""},
		CreatedBy:         createdBy,
		CreatedAt:         word.CreatedAt,
		UpdatedAt:         word.UpdatedAt,
	}
}

func filterStrings(f domain.WordFilter) (wc, src, stat, isRoot string) {
	if f.WordClass != nil {
		wc = string(*f.WordClass)
	}
	if f.Source != nil {
		src = string(*f.Source)
	}
	if f.Status != nil {
		stat = string(*f.Status)
	}
	if f.IsRoot != nil {
		if *f.IsRoot {
			isRoot = "true"
		} else {
			isRoot = "false"
		}
	}
	return
}

func listParams(f domain.WordFilter, page, perPage int) q.ListWordsParams {
	wc, src, stat, isRoot := filterStrings(f)
	sort := f.Sort
	if sort == "" {
		sort = "alphabetical"
	}
	return q.ListWordsParams{
		Column1: wc,
		Column2: isRoot,
		Column3: src,
		Column4: stat,
		Column5: sort,
		Limit:   int32(perPage),
		Offset:  int32((page - 1) * perPage),
	}
}

func isUniqueViolation(err error) bool {
	return err != nil && (containsStr(err.Error(), "23505") || containsStr(err.Error(), "unique") || containsStr(err.Error(), "words_unique"))
}

func containsStr(s, sub string) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
