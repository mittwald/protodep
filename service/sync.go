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

	newdeps := make([]dependency.ProtoDepDependency, 0, len(protodep.Dependencies))
	protodepDir := filepath.Join(s.userHomeDir, ".protodep")

	outdir := filepath.Join(s.outputRootDir, protodep.ProtoOutdir)
	if err := os.RemoveAll(outdir); err != nil {
		return err
	}

	var authProvider helper.AuthProvider
	for _, dep := range protodep.Dependencies {

		repoURL, err := url.Parse("https://" + dep.Target)
		if err != nil {
			return err
		}

		repoHostnameWithScheme := repoURL.Scheme + "://" + repoURL.Hostname()
		isSshGitRepo := helper.GitConfig(repoHostnameWithScheme)
		if len(isSshGitRepo) > 0 {
			dep.Target = isSshGitRepo
			authProvider = s.authProviderSSH
		} else {
			authProvider = s.authProviderHTTPS
		}

		logger.Info("using %v as authentication for repo %s", reflect.TypeOf(authProvider), dep.Target)
		gitRepo := repository.NewGitRepository(protodepDir, dep, authProvider)

		repo, err := gitRepo.Open()
		if err != nil {
			return err
		}

		sources := make([]protoResource, 0)

		protoRootDir := gitRepo.ProtoRootDir()
		_ = filepath.Walk(protoRootDir, func(path string, info os.FileInfo, err error) error {
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

		for _, s := range sources {
			outpath := filepath.Join(outdir, dep.Path, s.relativeDest)

			content, err := ioutil.ReadFile(s.source)
			if err != nil {
				return err
			}

			if err := helper.WriteFileWithDirectory(outpath, content, 0644); err != nil {
				return err
			}
		}

		newdeps = append(newdeps, dependency.ProtoDepDependency{
			Target:   repo.Dep.Target,
			Branch:   repo.Dep.Branch,
			Revision: repo.Hash,
			Path:     repo.Dep.Path,
			Ignores:  repo.Dep.Ignores,
		})
	}

	newProtodep := dependency.ProtoDep{
		ProtoOutdir:  protodep.ProtoOutdir,
		Dependencies: newdeps,
	}

	if dep.IsNeedWriteLockFile() {
		if err := helper.WriteToml("protodep.lock", newProtodep); err != nil {
			return err
		}
	}

	return nil
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
