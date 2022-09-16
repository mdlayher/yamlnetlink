// Command yamlnetlink-go generates Go code from YAML netlink specifications.
package main

import (
	"flag"
	"log"
	"os"

	"github.com/mdlayher/yamlnetlink"
)

func main() {
	log.SetFlags(0)

	pFlag := flag.String("p", "", "optional: specify a package name for the generated code (default: use YAML netlink spec name)")
	flag.Parse()

	path := flag.Arg(0)
	if path == "" {
		log.Fatal("must specify a YAML netlink file:\n$ yamlnetlink-go nlctrl.yaml")
	}

	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	s, err := yamlnetlink.Parse(f)
	if err != nil {
		log.Fatalf("failed to parse YAML netlink file: %v", err)
	}
	_ = f.Close()

	code, err := yamlnetlink.Generate(s, &yamlnetlink.Config{Package: *pFlag})
	if err != nil {
		log.Fatalf("failed to generate code: %v", err)
	}

	_, _ = os.Stdout.Write(code)
}
