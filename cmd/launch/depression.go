package launch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Depression struct{}

func (d *Depression) String() string {
	return "depression.ai"
}

func (d *Depression) Run(model string, models []LaunchModel, args []string) error {
	bin, err := findOrInstallDepression()
	if err != nil {
		return err
	}

	launchArgs := append([]string{}, args...)
	if model != "" {
		launchArgs = append(launchArgs, "--provider", "ollama", "--model", model)
	}

	cmd := exec.Command(bin, launchArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}

func findOrInstallDepression() (string, error) {
	if bin, err := exec.LookPath("depression"); err == nil {
		return bin, nil
	}

	if venv := os.Getenv("VIRTUAL_ENV"); venv != "" {
		if bin, err := depressionPath(venv); err == nil {
			return bin, nil
		}
	}

	venv, err := depressionManagedVenv()
	if err != nil {
		return "", err
	}
	if bin, err := depressionPath(venv); err == nil {
		return bin, nil
	}

	python := ""
	for _, name := range []string{"python3", "python"} {
		if path, err := exec.LookPath(name); err == nil {
			python = path
			break
		}
	}
	if python == "" {
		return "", fmt.Errorf("Python was not found; install Python 3.10 or newer")
	}

	if err := os.MkdirAll(filepath.Dir(venv), 0o755); err != nil {
		return "", fmt.Errorf("cannot create depression.ai environment directory: %w", err)
	}

	if _, err := os.Stat(filepath.Join(venv, "pyvenv.cfg")); err != nil {
		create := exec.Command(python, "-m", "venv", venv)
		create.Stdout, create.Stderr = os.Stdout, os.Stderr
		if err := create.Run(); err != nil {
			return "", fmt.Errorf("cannot create Python virtual environment: %w", err)
		}
	}

	if err := installDepression(venvPython(venv)); err != nil {
		return "", err
	}
	return depressionPath(venv)
}

func depressionInstalled() bool {
	if _, err := exec.LookPath("depression"); err == nil {
		return true
	}
	if venv := os.Getenv("VIRTUAL_ENV"); venv != "" {
		if _, err := depressionPath(venv); err == nil {
			return true
		}
	}
	venv, err := depressionManagedVenv()
	if err != nil {
		return false
	}
	_, err = depressionPath(venv)
	return err == nil
}

func ensureDepressionInstalled() error {
	_, err := findOrInstallDepression()
	return err
}

func depressionManagedVenv() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("cannot locate user cache directory: %w", err)
	}
	return filepath.Join(cache, "ollama", "depression-ai"), nil
}

func venvPython(venv string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venv, "Scripts", "python.exe")
	}
	return filepath.Join(venv, "bin", "python")
}

func installDepression(python string) error {
	cmd := exec.Command(python, "-m", "pip", "install", "--upgrade", "depression.ai")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not install depression.ai: %w", err)
	}
	return nil
}

func depressionPath(venv string) (string, error) {
	dir, name := "bin", "depression"
	if runtime.GOOS == "windows" {
		dir, name = "Scripts", "depression.exe"
	}
	path := filepath.Join(venv, dir, name)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("depression executable not found at %s", path)
	}
	return path, nil
}
