package main

import (
	"os"
	"os/exec"
	"path/filepath"
)

const dist_dir = "build"

var build_args []string = []string{
	"build",
	"-trimpath",
	"-ldflags=-s -w",
}

func get_build_cmd(args ...string) *exec.Cmd {
	args_ := append(build_args, args...)
	cmd := exec.Command("go", args_...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd
}
func build_server(arch string) error {
	pkg_name := "cmd/scrcpy-go-server"

	pkg, err := filepath.Abs(pkg_name)
	if err != nil {
		return err
	}
	dist, err := filepath.Abs(filepath.Join(dist_dir, filepath.Base(pkg)))
	if err != nil {
		return err
	}

	cmd := get_build_cmd("-o", dist, pkg)
	cmd.Env = append(cmd.Env, "GOOS=linux", "GOARCH="+arch)
	return cmd.Run()
}

func main() {
	err := build_server("arm64")
	if err != nil {
		panic(err)
	}
}
