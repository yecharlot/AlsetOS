// alset-desktop: single entry — builds on alset-desktop-bridge flags.
// Usage: alset-desktop -web ../Alset-LISPAI-Runtime/web
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Prefer same process: re-exec bridge with defaults if invoked as alset-desktop
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	root := dir
	// walk up for module root
	for d := dir; d != "/" && d != "."; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			root = d
			break
		}
	}
	bridge := filepath.Join(root, "bin", "alset-desktop-bridge")
	if _, err := os.Stat(bridge); err != nil {
		// run via go when developing
		fmt.Println("Alset Desktop — compile bridge first:")
		fmt.Println("  go build -o bin/alset-desktop-bridge ./cmd/alset-desktop-bridge")
		fmt.Println("  ./bin/alset-desktop-bridge -shell desktop/shell -web <LISPAI/web>")
		os.Exit(1)
	}
	args := append([]string{}, os.Args[1:]...)
	hasShell, hasWeb := false, false
	for _, a := range args {
		if a == "-shell" {
			hasShell = true
		}
		if a == "-web" {
			hasWeb = true
		}
	}
	if !hasShell {
		args = append(args, "-shell", filepath.Join(root, "desktop", "shell"))
	}
	if !hasWeb {
		cand := filepath.Join(root, "..", "Alset-LISPAI-Runtime", "web")
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			args = append(args, "-web", cand)
		}
	}
	cmd := exec.Command(bridge, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
