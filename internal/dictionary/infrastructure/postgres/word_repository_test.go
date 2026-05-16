package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	repo "github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/infrastructure/postgres"
	"github.com/iqbaleff214/kamus-banjar-api-2/testutil"
)

func TestPostgresWordRepository_CreateAndFind(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresWordRepository(pool)
	ctx := context.Background()

	word, err := domain.NewWord("urang", domain.WordClassNomina, domain.DialectHulu)
	require.NoError(t, err)
	require.NoError(t, r.Create(ctx, word))

	found, err := r.FindByID(ctx, word.ID)
	require.NoError(t, err)
	assert.Equal(t, "urang", found.Banjar)
}

func TestPostgresWordRepository_FindAll(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresWordRepository(pool)
	ctx := context.Background()

	word, _ := domain.NewWord("banar", domain.WordClassAdjektiva, domain.DialectHulu)
	require.NoError(t, r.Create(ctx, word))

	words, total, err := r.FindAll(ctx, domain.WordFilter{}, 1, 20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.NotEmpty(t, words)
}

func TestPostgresWordRepository_Search(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresWordRepository(pool)
	ctx := context.Background()

	word, _ := domain.NewWord("inya", domain.WordClassNomina, domain.DialectHulu)
	require.NoError(t, r.Create(ctx, word))

	results, total, err := r.Search(ctx, "inya", domain.WordFilter{}, 1, 20)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)
	assert.NotEmpty(t, results)
}

func TestPostgresWordRepository_Update(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresWordRepository(pool)
	ctx := context.Background()

	word, _ := domain.NewWord("kawa", domain.WordClassVerba, domain.DialectHulu)
	require.NoError(t, r.Create(ctx, word))

	require.NoError(t, word.AddDefinition("mampu/bisa", 1))
	require.NoError(t, r.Update(ctx, word))

	found, err := r.FindByID(ctx, word.ID)
	require.NoError(t, err)
	assert.Len(t, found.Definitions, 1)
}

func TestPostgresWordRepository_SoftDelete(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresWordRepository(pool)
	ctx := context.Background()

	word, _ := domain.NewWord("hancap", domain.WordClassAdjektiva, domain.DialectHulu)
	require.NoError(t, r.Create(ctx, word))

	require.NoError(t, r.SoftDelete(ctx, word.ID))

	found, err := r.FindByID(ctx, word.ID)
	require.NoError(t, err)
	assert.True(t, found.IsDeleted())
}

func TestPostgresWordRepository_FindByID_NotFound(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	r := repo.NewPostgresWordRepository(pool)
	ctx := context.Background()

	word, _ := domain.NewWord("tmp", domain.WordClassNomina, domain.DialectHulu)
	_, err := r.FindByID(ctx, word.ID)
	assert.Error(t, err)
}
