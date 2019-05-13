package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mittwald/protodep/dependency"
	"github.com/mittwald/protodep/helper"
	"github.com/mittwald/protodep/logger"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type GitRepository interface {
	Open() (*OpenedRepository, error)
	ProtoRootDir() string
}

type GitHubRepository struct {
	protodepDir  string
	dep          dependency.ProtoDepDependency
	authProvider helper.AuthProvider
}

type OpenedRepository struct {
	Repository *git.Repository
	Dep        dependency.ProtoDepDependency
	Hash       string
}

func NewGitRepository(protodepDir string, dep dependency.ProtoDepDependency, authProvider helper.AuthProvider) GitRepository {
	return &GitHubRepository{
		protodepDir:  protodepDir,
		dep:          dep,
		authProvider: authProvider,
	}
}

func (r *GitHubRepository) fetchRepository(repopath string) (*git.Repository, error) {
	reponame := r.dep.Repository()

	var rep *git.Repository

	if stat, err := os.Stat(repopath); err == nil && stat.IsDir() {
		spinner := logger.InfoWithSpinner("Getting in existing dir %s ", reponame)

		rep, err = git.PlainOpen(repopath)
		if err != nil {
			return nil, errors.Wrap(err, "open repository is failed")
		}
		spinner.Stop()

		fetchOpts := &git.FetchOptions{
			Auth: r.authProvider.AuthMethod(),
		}

		if err := rep.Fetch(fetchOpts); err != nil {
			if err != git.NoErrAlreadyUpToDate {
				return nil, errors.Wrap(err, "fetch repository is failed")
			}
		}
		spinner.Finish()

	} else {
		spinner := logger.InfoWithSpinner("Getting new Repo %s ", reponame)
		rep, err = git.PlainClone(repopath, false, &git.CloneOptions{
			Auth: r.authProvider.AuthMethod(),
			URL:  r.authProvider.GetRepositoryURL(reponame),
		})
		if err != nil {
			return nil, errors.Wrap(err, "clone repository is failed")
		}
		spinner.Finish()
	}
	return rep, nil
}

func (r *GitHubRepository) getCommitHashFromBranch(rep *git.Repository, branch string, wt *git.Worktree) (*git.Repository, error) {
	revision := r.dep.Revision
	if revision == "" {
		target, err := rep.Storer.Reference(plumbing.ReferenceName(fmt.Sprintf("refs/remotes/origin/%s", branch)))
		if err != nil {
			return nil, errors.Wrapf(err, "change branch to %s is failed", branch)
		}

		if err := wt.Checkout(&git.CheckoutOptions{Hash: target.Hash()}); err != nil {
			return nil, errors.Wrapf(err, "checkout to %s is failed", revision)
		}

		head := plumbing.NewHashReference(plumbing.HEAD, target.Hash())
		if err := rep.Storer.SetReference(head); err != nil {
			return nil, errors.Wrapf(err, "set head to %s is failed", branch)
		}
	} else {
		hash := plumbing.NewHash(revision)
		if err := wt.Checkout(&git.CheckoutOptions{Hash: hash}); err != nil {
			return nil, errors.Wrapf(err, "checkout to %s is failed", revision)
		}

		head := plumbing.NewHashReference(plumbing.HEAD, hash)
		if err := rep.Storer.SetReference(head); err != nil {
			return nil, errors.Wrapf(err, "set head to %s is failed", revision)
		}
	}
	return rep, nil
}

func (r *GitHubRepository) Open() (*OpenedRepository, error) {

	branch := "master"
	if r.dep.Branch != "" {
		branch = r.dep.Branch
	}

	reponame := r.dep.Repository()
	repopath := filepath.Join(r.protodepDir, reponame)

	rep, err := r.fetchRepository(repopath)
	if err != nil {
		return nil, err
	}

	wt, err := rep.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "get worktree is failed")
	}

	rep, err = r.getCommitHashFromBranch(rep, branch, wt)
	if err != nil {
		return nil, err
	}

	committer, err := rep.Log(&git.LogOptions{All: true})
	if err != nil {
		return nil, errors.Wrap(err, "get commit is failed")
	}

	current, err := committer.Next()
	if err != nil {
		return nil, errors.Wrap(err, "get commit current is failed")
	}

	return &OpenedRepository{
		Repository: rep,
		Dep:        r.dep,
		Hash:       current.Hash.String(),
	}, nil
}

func (r *GitHubRepository) ProtoRootDir() string {
	return filepath.Join(r.protodepDir, r.dep.Target)
}
