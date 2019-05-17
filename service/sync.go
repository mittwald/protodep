package service

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mittwald/protodep/dependency"
	"github.com/mittwald/protodep/helper"
	"github.com/mittwald/protodep/logger"
	"github.com/mittwald/protodep/repository"
)

type protoResource struct {
	source       string
	relativeDest string
}

type Sync interface {
	Resolve(forceUpdate bool) error
}

type SyncImpl struct {
	authProviderSSH   helper.AuthProvider
	authProviderHTTPS helper.AuthProvider
	userHomeDir       string
	targetDir         string
	outputRootDir     string
}

func NewSync(authProviderSSH, authProviderHTTPS helper.AuthProvider, userHomeDir string, targetDir string, outputRootDir string) Sync {
	return &SyncImpl{
		authProviderSSH:   authProviderSSH,
		authProviderHTTPS: authProviderHTTPS,
		userHomeDir:       userHomeDir,
		targetDir:         targetDir,
		outputRootDir:     outputRootDir,
	}
}

func (s *SyncImpl) Resolve(forceUpdate bool) error {

	dep := dependency.NewDependency(s.targetDir, forceUpdate)
	protodep, err := dep.Load()

	if err != nil {
		return err
	}

	outdir := filepath.Join(s.outputRootDir, protodep.ProtoOutdir)

	if err := os.RemoveAll(outdir); err != nil {
		return err
	}

	newdeps, err := s.getNewDeps(protodep, outdir)
	if err != nil {
		return err
	}

	newProtodep := dependency.ProtoDep{
		ProtoOutdir:  protodep.ProtoOutdir,
		Dependencies: *newdeps,
	}

	if dep.IsNeedWriteLockFile() {
		if err := helper.WriteToml("protodep.lock", newProtodep); err != nil {
			return err
		}
	}

	return nil
}

func (s *SyncImpl) getSources(gitRepo repository.GitRepository, dep *dependency.ProtoDepDependency) ([]protoResource, error) {

	sources := make([]protoResource, 0)

	protoRootDir := gitRepo.ProtoRootDir()

	err := filepath.Walk(protoRootDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".proto") {
				if s.isIgnorePath(protoRootDir, path, dep.Ignores) {
					logger.Info("skipped %s due to ignore setting", path)
				} else {
					sources = append(sources, protoResource{
						source:       path,
						relativeDest: strings.Replace(path, protoRootDir, "", -1),
					})
				}
			}
			return nil
		})
	return sources, err
}

func (s *SyncImpl) getAuthProvider(rewrittenGitRepo string, repoURL *url.URL, dep *dependency.ProtoDepDependency, bareDepRepo string) (helper.AuthProvider, error) {
	var authProvider helper.AuthProvider

	if len(rewrittenGitRepo) > 0 {
		logger.Info("found rewrite in gitconfig for '%s' ...", bareDepRepo)

		// port := strings.Split(rewrittenGitRepo, ":")[1]

		// if len(port) > 0 {
		// 	rewrittenGitRepo = rewrittenGitRepo[0:strings.Index(rewrittenGitRepo, ":")]
		// }

		rewrittenGitRepoURL, err := url.Parse(rewrittenGitRepo)
		if err != nil {
			return nil, err
		}

		dep.Target = rewrittenGitRepo + repoURL.Path

		logger.Info("... rewriting to '%s'", dep.Target)

		if rewrittenGitRepoURL.Scheme == "ssh" {
			authProvider = s.authProviderSSH
		} else {
			authProvider = s.authProviderHTTPS
		}
	} else {
		authProvider = s.authProviderHTTPS
	}
	return authProvider, nil
}

func (s *SyncImpl) getNewDeps(protodep *dependency.ProtoDep, outdir string) (*[]dependency.ProtoDepDependency, error) {

	newdeps := make([]dependency.ProtoDepDependency, 0, len(protodep.Dependencies))

	var protoDepCachePath string
	protoDepCachePath = os.Getenv("PROTODEP_CACHE_PATH")
	if len(protoDepCachePath) <= 0 {
		protoDepCachePath = filepath.Join(s.userHomeDir, ".protodep")
	}

	for _, dep := range protodep.Dependencies {

		depRepoURL, err := url.Parse("https://" + dep.Target)
		if err != nil {
			logger.Error("failed to parse dep Target '%s'", dep.Target)
			return nil, err
		}

		bareDepHostname := depRepoURL.Hostname()
		bareDepRepoPath := strings.TrimPrefix(depRepoURL.Path, "/")
		bareDepRepo := bareDepHostname + "/" + bareDepRepoPath

		repoURL, err := url.Parse("https://" + bareDepRepo)

		if err != nil {
			return nil, err
		}

		repoHostnameWithScheme := repoURL.Scheme + "://" + repoURL.Hostname()

		r, err := helper.LoadGitConfigFileFromHome()
		if err != nil {
			logger.Error("%v", err)
		}
		rewrittenGitRepos, err := helper.GitConfig(repoHostnameWithScheme, r)
		if err != nil {
			return nil, err
		}

		var authProvider helper.AuthProvider
		for i, v := range rewrittenGitRepos {
			authProvider, err = s.getAuthProvider(v, repoURL, &dep, bareDepRepo)
			if err != nil {
				logger.Info("Try %d failed : %s", i+1, v)
				continue
			}
		}
		if authProvider == nil {
			authProvider, err = s.getAuthProvider("", repoURL, &dep, bareDepRepo)
			if err != nil {
				return nil, err
			}
		}

		logger.Info("using %v as authentication for repo %s", reflect.TypeOf(authProvider), dep.Target)
		gitRepo := repository.NewGitRepository(protoDepCachePath, dep, authProvider)

		repo, err := gitRepo.Open()
		if err != nil {
			return nil, err
		}

		sources, _ := s.getSources(gitRepo, &dep)

		for _, s := range sources {
			outpath := filepath.Join(outdir, dep.Path, s.relativeDest)

			content, err := ioutil.ReadFile(s.source)
			if err != nil {
				return nil, err
			}

			if err := helper.WriteFileWithDirectory(outpath, content, 0644); err != nil {
				return nil, err
			}
		}

		newdeps = append(newdeps, dependency.ProtoDepDependency{
			Target:   bareDepRepo,
			Branch:   repo.Dep.Branch,
			Revision: repo.Hash,
			Path:     repo.Dep.Path,
			Ignores:  repo.Dep.Ignores,
		})
	}

	return &newdeps, nil
}

func (s *SyncImpl) isIgnorePath(protoRootDir string, target string, ignores []string) bool {

	for _, ignore := range ignores {
		pathPrefix := filepath.Join(protoRootDir, ignore)
		if strings.HasPrefix(target, pathPrefix) {
			return true
		}
	}

	return false
}
