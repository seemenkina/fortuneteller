package repository

import (
	"context"
	"testing"

	"fortuneteller/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookfs_ListBooks(t *testing.T) {
	repo := NewBookFileSystem("testdata/books")
	books, err := repo.ListBooks(context.TODO())
	require.NoError(t, err)
	assert.ElementsMatch(t, books, []data.Book{
		{"b1", 5},
	})
}

func TestBookfs_FindRowInBook(t *testing.T) {
	repo := NewBookFileSystem("testdata/books")
	str, err := repo.FindRowInBook(context.TODO(), "b1", 3)
	require.NoError(t, err)
	assert.Equal(t, str, "4")
}
