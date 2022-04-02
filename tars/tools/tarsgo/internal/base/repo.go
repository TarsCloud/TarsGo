package base

import (
	"context"
	"fmt"
	"github.com/TarsCloud/TarsGo/tars/tools/tarsgo/internal/consts"
	"os"
	"os/exec"
	"path"
	"strings"
)

// Repo is git repository manager.
type Repo struct {
	url     string
	home    string
	branch  string
	demoDir string
}

// NewRepo new a repository manager.
func NewRepo(url string, branch string, demoDir string) *Repo {
	var start int
	start = strings.Index(url, "//")
	if start == -1 {
		start = strings.Index(url, ":") + 1
	} else {
		start += 2
	}
	end := strings.LastIndex(url, "/")
	return &Repo{
		url:     url,
		home:    TarsGoHomeWithDir("repo/" + url[start:end]),
		branch:  branch,
		demoDir: demoDir,
	}
}

// Path returns the repository cache path.
func (r *Repo) Path() string {
	start := strings.LastIndex(r.url, "/")
	end := strings.LastIndex(r.url, ".git")
	if end == -1 {
		end = len(r.url)
	}
	var branch string
	if r.branch == "" {
		branch = "@main"
	} else {
		branch = "@" + r.branch
	}
	return path.Join(r.home, r.url[start+1:end]+branch)
}

// Pull fetch the repository from remote url.
func (r *Repo) Pull(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "symbolic-ref", "HEAD")
	cmd.Dir = r.Path()
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	cmd = exec.CommandContext(ctx, "git", "pull")
	cmd.Dir = r.Path()
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		return err
	}
	return err
}

// Clone clones the repository to cache path.
func (r *Repo) Clone(ctx context.Context) error {
	if _, err := os.Stat(r.Path()); !os.IsNotExist(err) {
		return r.Pull(ctx)
	}
	var cmd *exec.Cmd
	if r.branch == "" {
		cmd = exec.CommandContext(ctx, "git", "clone", r.url, r.Path())
	} else {
		cmd = exec.CommandContext(ctx, "git", "clone", "-b", r.branch, r.url, r.Path())
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// CopyTo copies the repository to project path.
func (r *Repo) CopyTo(ctx context.Context, p *Project, to string, demoDir string, ignores []string) error {
	if err := r.Clone(ctx); err != nil {
		return err
	}
	replaces := []string{
		"_APP_", p.App,
		"_SERVER_", p.Server,
		"_SERVANT_", p.Servant,
		"_MODULE_", p.GoModuleName,
		// makefile
		"$(foreach path,$(libpath),$(eval -include $(path)/src/github.com/TarsCloud/TarsGo/tars/makefile.tars.gomod))", "-include scripts/makefile.tars.gomod",
		// CMakeLists.txt
		"${GOPATH}/src/github.com/TarsCloud/TarsGo/", "${CMAKE_CURRENT_SOURCE_DIR}/",
	}
	err := CopyDir(path.Join(r.Path(), "tars", "tools", r.demoDir), to, replaces, ignores)
	if err != nil {
		return err
	}
	err = CopyDir(path.Join(r.Path(), "tars", "tools", "debugtool"), path.Join(to, "debugtool"), replaces, ignores)
	if demoDir == consts.MakeDemoDir {
		_ = os.MkdirAll(path.Join(to, "scripts"), 0755)
		_ = CopyFile(path.Join(r.Path(), "tars", "makefile.tars.gomod"), path.Join(to, "scripts", "makefile.tars.gomod"), replaces)
	} else {
		replaces = []string{
			// cmake
			"${GOPATH}/src/github.com/TarsCloud/TarsGo/", "${CMAKE_CURRENT_SOURCE_DIR}/",
		}
		_ = CopyDir(path.Join(r.Path(), "cmake"), path.Join(to, "cmake"), replaces, ignores)
	}
	return err
}
