package git

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type File struct {
	Path  string
	IsDir bool
}

type gitFS interface {
	ListFiles() ([]File, error)
	Read(path string) (body string, err error)
	Write(path, body string) error
	MkDir(path string) error
	Delete(path string) error
}

type Repo struct {
	repo *git.Repository
	path string
}

var _ gitFS = new(Repo)

func InitRepo(path string) (*Repo, error) {
	r, err := git.PlainInit(path, false)
	if err != nil {
		return nil, err
	}
	return &Repo{repo: r, path: path}, nil
}

func OpenRepo(path string) (*Repo, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &Repo{repo: r, path: path}, nil
}

func (r *Repo) ListFiles() ([]File, error) {
	files := []File{}
	if err := filepath.WalkDir(r.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if r.path == path {
			return nil
		}
		if path == "" {
			return nil
		}
		if d.Name() == ".git" {
			return filepath.SkipDir
		}
		rpath, _ := filepath.Rel(r.path, path)
		files = append(files, File{
			Path:  rpath,
			IsDir: d.IsDir(),
		})
		return nil
	}); err != nil {
		return nil, err
	}
	return files, nil
}
func (r *Repo) Read(path string) (body string, err error) {
	b, err := os.ReadFile(r.join(path))
	return string(b), err
}

func (r *Repo) commitChange(msg string, path string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(path, "./") {
		path = "./" + path
	}
	if _, err := w.Add(path); err != nil {
		return err
	}
	if _, err := w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Steady",
			Email: "steady@steady",
			When:  time.Now(),
		},
	}); err != nil {
		return err
	}
	return nil
}

func (r *Repo) Write(path, body string) error {
	if _, err := os.Stat(r.join(path)); os.IsNotExist(err) {
		return err
	}
	if err := os.WriteFile(r.join(path), []byte(body), 0666); err != nil {
		return err
	}
	if err := r.commitChange(
		fmt.Sprintf("edited %s", path),
		path,
	); err != nil {
		return err
	}
	return nil
}

func (r *Repo) join(path string) string {
	if r.path == "" {
		panic("repo is not initialized")
	}
	joined := filepath.Join(r.path, path)
	// Don't allow accessing directories above the parent dir
	if !strings.HasPrefix(joined, r.path) {
		return r.path
	}
	return joined
}
func (r *Repo) Create(path string) error {
	if _, err := os.Stat(r.join(path)); !os.IsNotExist(err) {
		return fmt.Errorf("a file already exists at this path")
	}
	f, err := os.OpenFile(r.join(path), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}
	if err := r.commitChange(
		fmt.Sprintf("edited %s", path),
		path,
	); err != nil {
		return err
	}
	return nil
}
func (r *Repo) MkDir(path string) error {
	return os.Mkdir(r.join(path), 0777)
}
func (r *Repo) Delete(path string) error {
	if err := os.Remove(r.join(path)); err != nil {
		return err
	}
	if err := r.commitChange(fmt.Sprintf("deleted %s", path), path); err != nil {
		return err
	}
	return nil
}
