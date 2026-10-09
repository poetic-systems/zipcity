package gitcache

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

func PrepareFS() (fs.FS, error) {
	targetModule := "github.com/poetic-systems/zipcity"
	repoURL := fmt.Sprintf("https://%s", targetModule)
	localPath := "./cached-sparse-zipcity-repo"
	filterfspath := path.Join("generated", "embedded_filter")
	targetSubPath := path.Join(filterfspath, "data")
	version := ""

	var repo *git.Repository
	var err error

	// Try opening the repository if it already exists on disk
	repo, err = git.PlainOpen(localPath)
	if err == git.ErrRepositoryNotExists {

		buildInfo, ok := debug.ReadBuildInfo()
		if !ok {
			return nil, fmt.Errorf("Failed to read build info")
		}

		for _, dep := range buildInfo.Deps {
			if dep.Path == targetModule {
				version = dep.Version
				break
			}
		}

		if len(version) == 0 {
			if testing.Testing() {
				_, filename, _, ok := runtime.Caller(0)
				if !ok {
					return nil, fmt.Errorf("failed to get path for gitcache while testing")
				}
				fmt.Printf("Failed to get version info while testing. Attempting to find git repo for %q.\n", filename)
				repo, err = git.PlainOpenWithOptions(filename, &git.PlainOpenOptions{
					DetectDotGit: true,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to open repository while testing: %w", err)
				}
			} else {
				return nil, fmt.Errorf("Unable to determine version of %s dependency", targetModule)
			}
		}

		if repo == nil {
			versionRef := plumbing.NewTagReferenceName(version)

			// No existing cache. Clone the repository to disk for the first time
			repo, err = git.PlainClone(localPath, false, &git.CloneOptions{
				URL:           repoURL,
				Depth:         1,          // --depth 1 (only the latest commit history)
				SingleBranch:  true,       // --single-branch (skips downloading other branches)
				NoCheckout:    true,       // --no-checkout (prevents writing any files to disk initially)
				ReferenceName: versionRef, // Explicitly targeting our version tag reference
			})
			if err != nil {
				return nil, fmt.Errorf("failed to clone repository: %w", err)
			}
			w, err := repo.Worktree()
			if err != nil {
				return nil, err
			}
			err = w.Checkout(&git.CheckoutOptions{
				Branch:                    versionRef,
				SparseCheckoutDirectories: []string{targetSubPath}, // Restricts checkout to this sub-path
				Force:                     true,
			})
			if err != nil {
				return nil, fmt.Errorf("failed sparse checkout: %w", err)
			}
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to open existing repository: %w", err)
	} else {
		// Cache hit. Use the existing local data
	}

	storer, ok := repo.Storer.(*filesystem.Storage)
	if !ok {
		return nil, fmt.Errorf("repository storage is not *filesystem.Storage")
	}

	fs := storer.Filesystem()
	reporoot, _ := strings.CutSuffix(fs.Root(), ".git")
	fmt.Printf("Using git repo root: %q\n", reporoot)

	if !strings.HasSuffix(reporoot, "zipcity/") {
		return nil, fmt.Errorf("repository is not zipcity")
	}

	datafilepath := path.Join(reporoot, filterfspath)
	fmt.Printf("Using git based filter filesystem path: %q\n", datafilepath)

	checkoutFS := os.DirFS(datafilepath)

	return checkoutFS, nil
}
