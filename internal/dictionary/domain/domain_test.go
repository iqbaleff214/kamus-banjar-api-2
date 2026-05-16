package domain_test

import (
	"strings"
	"testing"

	"github.com/iqbaleff214/kamus-banjar-api-2/internal/dictionary/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- value objects ---

func TestWordClass_Valid(t *testing.T) {
	valid := []domain.WordClass{
		domain.WordClassNomina, domain.WordClassVerba, domain.WordClassAdjektiva,
		domain.WordClassAdverbia, domain.WordClassPartikel,
		domain.WordClassPribahasa, domain.WordClassKiasan,
	}
	for _, wc := range valid {
		assert.NoError(t, wc.Validate(), "expected %q to be valid", wc)
	}
}

func TestWordClass_Invalid(t *testing.T) {
	err := domain.WordClass("x").Validate()
	assert.ErrorIs(t, err, domain.ErrInvalidWordClass)
}

func TestDialect_OnlyHulu(t *testing.T) {
	assert.NoError(t, domain.DialectHulu.Validate())
	err := domain.Dialect("kuala").Validate()
	assert.ErrorIs(t, err, domain.ErrInvalidDialect)
}

func TestSource_Valid(t *testing.T) {
	assert.NoError(t, domain.SourceSeeded.Validate())
	assert.NoError(t, domain.SourceContributed.Validate())
	assert.NoError(t, domain.SourceAIGenerated.Validate())
}

func TestSource_Invalid(t *testing.T) {
	err := domain.Source("unknown").Validate()
	assert.ErrorIs(t, err, domain.ErrInvalidSource)
}

func TestWordStatus_Valid(t *testing.T) {
	assert.NoError(t, domain.WordStatusActive.Validate())
	assert.NoError(t, domain.WordStatusDeprecated.Validate())
}

// --- definition ---

func TestDefinition_MeaningTooLong(t *testing.T) {
	long := strings.Repeat("a", 2001)
	_, err := domain.NewDefinition(long, 1, domain.SourceSeeded)
	assert.ErrorIs(t, err, domain.ErrMeaningTooLong)
}

func TestDefinition_NetScore(t *testing.T) {
	def, err := domain.NewDefinition("ayah", 1, domain.SourceSeeded)
	require.NoError(t, err)
	def.Upvotes = 5
	def.Downvotes = 2
	assert.Equal(t, 3, def.NetScore())
}

func TestDefinition_Valid(t *testing.T) {
	def, err := domain.NewDefinition("ayah", 1, domain.SourceSeeded)
	require.NoError(t, err)
	assert.Equal(t, "ayah", def.Meaning)
	assert.Equal(t, 1, def.SortOrder)
	assert.Equal(t, domain.SourceSeeded, def.Source)
	assert.NotEqual(t, "", def.ID.String())
}

// --- word aggregate ---

func TestNewWord_Valid(t *testing.T) {
	w, err := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	require.NoError(t, err)
	assert.Equal(t, "abah", w.Banjar)
	assert.Equal(t, domain.WordClassNomina, w.WordClass)
	assert.Equal(t, domain.DialectHulu, w.Dialect)
	assert.Equal(t, domain.WordStatusActive, w.Status)
	assert.Equal(t, domain.SourceSeeded, w.Source)
	assert.True(t, w.IsRoot)
	assert.Equal(t, 1, w.HomonymNumber)
}

func TestNewWord_EmptyBanjar(t *testing.T) {
	_, err := domain.NewWord("", domain.WordClassNomina, domain.DialectHulu)
	assert.ErrorIs(t, err, domain.ErrEmptyBanjar)
}

func TestNewWord_InvalidWordClass(t *testing.T) {
	_, err := domain.NewWord("abah", domain.WordClass("xyz"), domain.DialectHulu)
	assert.ErrorIs(t, err, domain.ErrInvalidWordClass)
}

func TestNewWord_InvalidDialect(t *testing.T) {
	_, err := domain.NewWord("abah", domain.WordClassNomina, domain.Dialect("kuala"))
	assert.ErrorIs(t, err, domain.ErrInvalidDialect)
}

func TestAddDefinition_TooLong(t *testing.T) {
	w, _ := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	err := w.AddDefinition(strings.Repeat("a", 2001), 1)
	assert.ErrorIs(t, err, domain.ErrMeaningTooLong)
}

func TestAddDefinition_Valid(t *testing.T) {
	w, _ := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	err := w.AddDefinition("ayah", 1)
	require.NoError(t, err)
	require.Len(t, w.Definitions, 1)
	assert.Equal(t, "ayah", w.Definitions[0].Meaning)
}

func TestAddExample(t *testing.T) {
	w, _ := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	w.AddExample("inya kada barabah", "dia tidak berayah")
	require.Len(t, w.Examples, 1)
	assert.Equal(t, "inya kada barabah", w.Examples[0].BanjarSentence)
}

func TestDeprecateAndRestore(t *testing.T) {
	w, _ := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	assert.Equal(t, domain.WordStatusActive, w.Status)

	w.Deprecate()
	assert.Equal(t, domain.WordStatusDeprecated, w.Status)

	w.Restore()
	assert.Equal(t, domain.WordStatusActive, w.Status)
}

func TestSoftDelete(t *testing.T) {
	w, _ := domain.NewWord("abah", domain.WordClassNomina, domain.DialectHulu)
	assert.False(t, w.IsDeleted())

	w.SoftDelete()
	assert.True(t, w.IsDeleted())
	assert.NotNil(t, w.DeletedAt)
}
