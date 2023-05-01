package git_test

import (
	"testing"

	"github.com/maxmcd/steady/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	tmpDir := t.TempDir()
	repo, err := git.InitRepo(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	require.Error(t,
		repo.Write("./steady", "void"),
		"can't write to a file that doesn't exist")
	if err := repo.Create("steady"); err != nil {
		t.Fatal(err)
	}
	if err := repo.MkDir("stable"); err != nil {
		t.Fatal(err)
	}
	require.Error(t,
		repo.Write("./stable", "dir contents?!"),
		"can't write contents to a directory")

	if err := repo.Create("./stable/steady"); err != nil {
		t.Fatal(err)
	}
	require.Error(t,
		repo.Create("./stable/steady"),
		"can't create file twice")

	{
		contents := "rock steady"
		if err := repo.Write("./stable/steady", contents); err != nil {
			t.Fatal(err)
		}
		readContents, err := repo.Read("./stable/steady")
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, contents, readContents)
	}

	files, err := repo.ListFiles()
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, []git.File{
		{Path: "stable", IsDir: true},
		{Path: "stable/steady", IsDir: false},
		{Path: "steady", IsDir: false},
	}, files)

	require.Error(t, repo.Delete("stable"), "can't delete a directory unless it is empty")

	if err := repo.Delete("steady"); err != nil {
		t.Fatal(err)
	}

	{
		files, err := repo.ListFiles()
		if err != nil {
			t.Fatal(err)
		}
		require.Equal(t, []git.File{
			{Path: "stable", IsDir: true},
			{Path: "stable/steady", IsDir: false},
		}, files)
	}
	{
		repo, err := git.OpenRepo(tmpDir)
		if err != nil {
			t.Fatal(err)
		}
		if err := repo.Write("stable/steady", "updated"); err != nil {
			t.Fatal(err)
		}
		body, err := repo.Read("stable/steady")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "updated", body)
	}
}
