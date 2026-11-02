// Command priorart runs the prior-art MCP server over stdio.
// Subcommand: `priorart index <root> <outDir>` builds the capability index and exits.
package main

import (
	"fmt"
	"os"

	"github.com/berlinguyinca/priorart/internal/backends"
	"github.com/berlinguyinca/priorart/internal/backends/crates"
	"github.com/berlinguyinca/priorart/internal/backends/github"
	"github.com/berlinguyinca/priorart/internal/backends/grepapp"
	"github.com/berlinguyinca/priorart/internal/backends/librariesio"
	"github.com/berlinguyinca/priorart/internal/backends/npm"
	"github.com/berlinguyinca/priorart/internal/backends/ownlocal"
	"github.com/berlinguyinca/priorart/internal/config"
	"github.com/berlinguyinca/priorart/internal/core"
	"github.com/berlinguyinca/priorart/internal/index"
	"github.com/berlinguyinca/priorart/internal/mcp"
)

func main() {
	// `priorart index <root> <outDir>` builds the capability index and exits.
	if len(os.Args) >= 4 && os.Args[1] == "index" {
		if _, err := index.Build(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintln(os.Stderr, "priorart: index error:", err)
			os.Exit(1)
		}
		return
	}
	cfgPath := ""
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "priorart: config error:", err)
		os.Exit(1)
	}
	bs := []backends.Backend{
		ownlocal.New(cfg.OwnRoots, "rg"),
		npm.New("", nil),
		librariesio.New("", cfg.LibrariesIOKey, nil),
		crates.New("", nil),
		github.New("", cfg.GitHubToken, nil),
		grepapp.New("", nil),
	}
	c := core.New(cfg, bs)
	c.SetLedgerPath(cfg.LedgerPath)
	if err := mcp.New(c).Serve(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "priorart: serve error:", err)
		os.Exit(1)
	}
}
