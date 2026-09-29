// alset-pkg: list / install Alset-native packages (.alset.json), not TCZ.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yecharlot/AlsetOS/pkg/alsetpkg"
)

func main() {
	root := flag.String("root", "", "install root (default ~/.alset/pkg)")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("uso: alset-pkg list | install <file.alset.json> | show <name>")
		os.Exit(1)
	}
	base := *root
	if base == "" {
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".alset", "pkg")
	}
	_ = os.MkdirAll(base, 0o755)

	switch args[0] {
	case "list":
		entries, _ := os.ReadDir(base)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".alset.json") {
				fmt.Println(e.Name())
			}
		}
	case "install":
		if len(args) < 2 {
			fmt.Println("install requiere archivo")
			os.Exit(1)
		}
		m, err := alsetpkg.Load(args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		dst := filepath.Join(base, m.Name+".alset.json")
		if err := alsetpkg.Save(dst, m); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("instalado:", dst)
	case "show":
		if len(args) < 2 {
			os.Exit(1)
		}
		p := filepath.Join(base, args[1])
		if !strings.HasSuffix(p, ".alset.json") {
			p += ".alset.json"
		}
		b, err := os.ReadFile(p)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(b))
	default:
		fmt.Println("comando desconocido")
		os.Exit(1)
	}
}
