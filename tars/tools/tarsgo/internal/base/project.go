package base

import (
	"context"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"path"
)

// Project is a project template.
type Project struct {
	App          string
	Server       string
	Servant      string
	GoModuleName string
}

func NewProject(app, server, servant, goModuleName string) *Project {
	return &Project{
		App:          app,
		Server:       server,
		Servant:      servant,
		GoModuleName: goModuleName,
	}
}

// New new a project from remote repo.
func (p *Project) Create(ctx context.Context, dir string, layout string, branch string, demoDir string) error {
	to := path.Join(dir, p.Server)
	if _, err := os.Stat(to); !os.IsNotExist(err) {
		fmt.Printf("🚫 %s already exists\n", p.Server)
		override := false
		prompt := &survey.Confirm{
			Message: "📂 Do you want to override the folder ?",
			Help:    "Delete the existing folder and create the project.",
		}
		e := survey.AskOne(prompt, &override)
		if e != nil {
			return e
		}
		if !override {
			return err
		}
		os.RemoveAll(to)
	}

	if err := GoInstall("github.com/TarsCloud/TarsGo/tars/tools/tars2go"); err != nil {
		return err
	}

	fmt.Printf("🚀 Creating server %s.%s, layout repo is %s, please wait a moment.\n\n", p.App, p.Server, layout)
	repo := NewRepo(layout, branch, demoDir)
	if err := repo.CopyTo(ctx, p, to, demoDir, []string{".git"}); err != nil {
		return err
	}

	if err := os.Rename(path.Join(to, "Servant.tars"), path.Join(to, p.Servant+".tars")); err != nil {
		return err
	}

	if err := os.Rename(path.Join(to, "Servant_imp.go"), path.Join(to, p.Servant+"_imp.go")); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "go", "mod", "init", p.GoModuleName)
	cmd.Dir = to
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	Tree(to, dir)

	fmt.Printf("\n>>> Great！Done! You can jump in %s\n", color.GreenString(p.Server))
	fmt.Println(">>> Tips: After editing the Tars file, execute the following cmd to automatically generate golang files.")
	fmt.Printf(">>>       %s/bin/tars2go *.tars\n", os.Getenv("GOPATH"))

	fmt.Println(color.WhiteString("$ cd %s", p.Server))
	fmt.Println(color.WhiteString("$ ./start.sh"))
	fmt.Println("🤝 Thanks for using TarsGo")
	fmt.Println("📚 Tutorial: https://tarscloud.github.io/TarsDocs/")
	return nil
}
