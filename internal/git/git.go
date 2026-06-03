package git

import (
	"fmt"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
)

// CheckoutRepos checks out a list of git repositories to the specified destination directory. If a repository already exists in the destination, it pulls the latest changes from the origin remote.
func CheckoutRepos(repos []string, destDir string) error {
	for i, repoURL := range repos {
		dir := filepath.Join(destDir, fmt.Sprintf("repo_%d", i))
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			_, err := gogit.PlainClone(dir, false, &gogit.CloneOptions{
				URL:      repoURL,
				Progress: os.Stdout,
			})
			if err != nil {
				return err
			}
		} else {
			r, err := gogit.PlainOpen(dir)
			if err != nil {
				return err
			}
			w, err := r.Worktree()
			if err != nil {
				return err
			}
			err = w.Pull(&gogit.PullOptions{RemoteName: "origin", Progress: os.Stdout})
			if err != nil && err != gogit.NoErrAlreadyUpToDate {
				return err
			}
		}
	}
	return nil
}
