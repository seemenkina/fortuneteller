package repository

import (
	"testing"

	"fortuneteller/internal/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookfs_ListBooks(t *testing.T) {
	repo := NewBookFileSystem("testdata/books", "")
	books, err := repo.ListBooks()
	require.NoError(t, err)
	assert.ElementsMatch(t, books, []data.Book{
		{"b1", 5},
	})
}

func TestBookfs_FindRowInBook(t *testing.T) {
	repo := NewBookFileSystem("testdata/books", "")
	str, err := repo.FindRowInBook("b1", 3)
	require.NoError(t, err)
	assert.Equal(t, str, "4")
}

func TestBookfs_GetBookKey(t *testing.T) {
	repo := NewBookFileSystem("testdata/books", "")
	bk := repo.GetBookKey("b1")
	// require.NoError(t, err)
	c := bk.Encrypt([]byte("FY0363251JDF9IC02BPFX245C3FCD66="))
	t.Logf("encrypted %v", c)
	p, err := bk.Decrypt(c)
	require.NoError(t, err)
	t.Logf("decrypted %s", p)
}
