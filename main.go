package main

import (
	"flag"
	"fmt"
	_ "embed"
	"github.com/neovim/go-client/nvim"
	"os"
	"path/filepath"
	"strings"
	"strconv"
)

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
)

type nvcatCliFlags struct {
	lineNumbers *bool
	clean       *bool
	help        *bool
	version     *bool
}

type formatOpts struct {
	tab string
}

var cliFlags = nvcatCliFlags{
	lineNumbers: flag.Bool("n", false, "Show line numbers"),
	clean:       flag.Bool("clean", false, "Use a clean Neovim instance"),
	help:        flag.Bool("h", false, "Show help"),
	version:     flag.Bool("v", false, "Show version"),
}

//go:embed lua/init.lua
var initLuaScript string

var Version = "dev"

func main() {
	flag.Parse()

	if *cliFlags.version {
		fmt.Println("nvcat " + Version)
		os.Exit(0)
	}

	if len(flag.Args()) < 1 || *cliFlags.help {
		fmt.Println("Usage: nvcat [options] <file>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	filename := flag.Args()[0]
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(fileContent), "\n")

	var args = []string{"--embed", "--headless"}
	if *cliFlags.clean {
		args = append(args, "--clean")
	}
	args = append(args, "--cmd", fmt.Sprintf("let g:nvcat = '%s'", Version))
	vim, err := nvim.NewChildProcess(nvim.ChildProcessArgs(args...))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Neovim: %v\n", err)
		os.Exit(1)
	}
	defer vim.Close()

	var validNvim int
	vim.Call("has", &validNvim, "nvim-0.10")
	if validNvim != 1 {
		fmt.Fprintf(os.Stderr, "Error: nvcat requires nvim 0.10 or later\n")
		os.Exit(1)
	}

	vim.RegisterHandler("redraw", func(args []any) {})
	vim.RegisterHandler("Gui", func(args []any) {})
	vim.AttachUI(2 * len(lines), 80, map[string]any{})

	err = vim.ExecLua(initLuaScript, nil, nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading Lua script: %v\n", err)
	}

	var expandtab bool
	var tabstop int

	err = vim.OptionValue("expandtab", map[string]nvim.OptionValueScope{}, &expandtab)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting expandtab option: %v\n", err)
		os.Exit(1)
	}

	err = vim.OptionValue("tabstop", map[string]nvim.OptionValueScope{}, &tabstop)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting tabstop option: %v\n", err)
		os.Exit(1)
	}
	tab := strings.Repeat(" ", tabstop)

	err = vim.Command(fmt.Sprintf("edit %s", absFilename))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}

	processFile(vim, lines, formatOpts { tab: tab })
}

func processFile(vim *nvim.Nvim, lines []string, opts formatOpts) {
	numDigits := len(fmt.Sprintf("%d", len(lines)))
	for i, line := range lines {
		if *cliFlags.lineNumbers {
			fmt.Fprint(os.Stderr, Dim)
			fmt.Fprint(os.Stdout, fmt.Sprintf("%" + strconv.Itoa(numDigits) + "d ", i+1))
			fmt.Fprint(os.Stderr, Reset)
		}
		if len(line) == 0 {
			fmt.Fprintln(os.Stdout, "")
			continue
		}
		_, err := getHighlightedLine(vim, i, line, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting highlights for line %d: %v\n", i+1, err)
			fmt.Fprintln(os.Stdout, line)
		}
		fmt.Fprintln(os.Stdout, "")
	}
}

func rgbToAnsi(color uint64) string {
	r := uint8((color >> 16) & 0xFF)
	g := uint8((color >> 8) & 0xFF)
	b := uint8(color & 0xFF)
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

func getHighlightColor(hl map[string]any) (string, error) {
	var ansiCode strings.Builder

	if fg, ok := hl["fg"].(uint64); ok {
		if ansi := rgbToAnsi(fg); ansi != "" {
			ansiCode.WriteString(ansi)
		}
	}

	if bold, ok := hl["bold"].(bool); ok && bold == true {
		ansiCode.WriteString(Bold)
	}
	if italic, ok := hl["italic"].(bool); ok && italic == true {
		ansiCode.WriteString(Italic)
	}
	if underline, ok := hl["underline"].(bool); ok && underline == true {
		ansiCode.WriteString(Underline)
	}

	result := ansiCode.String()
	if result == "" {
		result = Reset
	}
	return result, nil
}

func getHighlightedLine(vim *nvim.Nvim, lineNum int, line string, opts formatOpts) (string, error) {
	var currentAnsi string

	for col := range len(line) {
		var hl map[string]any
		if line[col] == '\t' {
			fmt.Fprint(os.Stdout, opts.tab)
			continue
		}
		err := vim.ExecLua("return NvcatGetHl(...)", &hl, lineNum, col)
		if err != nil {
			if currentAnsi != "" {
				fmt.Fprint(os.Stderr, Reset)
				currentAnsi = ""
			}
			fmt.Fprint(os.Stdout, string(line[col]))
			continue
		}

		ansi, err := getHighlightColor(hl)
		if err != nil {
			if currentAnsi != "" {
				fmt.Fprint(os.Stderr, Reset)
				currentAnsi = ""
			}
			fmt.Fprint(os.Stdout, string(line[col]))
			continue
		}

		// Update ANSI escape sequence only if it changed
		if ansi != currentAnsi {
			if currentAnsi != "" {
				fmt.Fprint(os.Stderr, Reset)
			}
			fmt.Fprint(os.Stderr, ansi)
			currentAnsi = ansi
		}

		fmt.Fprint(os.Stdout, string(line[col]))
	}

	// Reset color at the end of the line
	if currentAnsi != "" {
		fmt.Fprint(os.Stderr, Reset)
	}

	return "", nil
}
