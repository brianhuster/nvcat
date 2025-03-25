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
	"bytes"
	"github.com/clipperhouse/uax29/words"
)

const (
	AnsiReset     = "\033[0m"
	AnsiBold      = "\033[1m"
	AnsiDim       = "\033[2m"
	AnsiItalic    = "\033[3m"
	AnsiUnderline = "\033[4m"
)

type nvcatCliFlags struct {
	number  *bool
	clean   *bool
	help    *bool
	version *bool
}

type formatOpts struct {
	tab string
}

var cliFlags = nvcatCliFlags{
	number: flag.Bool("n", false, "Show line numbers"),
	clean:       flag.Bool("clean", false, "Use a clean Neovim instance"),
	help:        flag.Bool("h", false, "Show help"),
	version:     flag.Bool("v", false, "Show version"),
}

//go:embed runtime/plugin/nvcat.lua
var LuaPluginScript string

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
	if bytes.Contains(fileContent, []byte{0}) {
		fmt.Fprintf(os.Stderr, "Binary files are not supported\n")
		os.Exit(1)
	}

	lines := strings.Split(string(fileContent), "\n")

	var args = []string{"--cmd", fmt.Sprintf("let g:nvcat = '%s'", Version), "--embed", "--headless"}
	if *cliFlags.clean {
		args = append(args, "--clean")
	}
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

	err = vim.ExecLua(LuaPluginScript, nil, nil)

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

	printLines(vim, lines, formatOpts { tab: tab })
}

func printLines(vim *nvim.Nvim, lines []string, opts formatOpts) {
	numDigits := len(fmt.Sprintf("%d", len(lines)))
	for i, line := range lines {
		if *cliFlags.number {
			fmt.Fprint(os.Stderr, AnsiDim)
			fmt.Fprint(os.Stdout, fmt.Sprintf("%" + strconv.Itoa(numDigits) + "d ", i+1))
			fmt.Fprint(os.Stderr, AnsiReset)
		}
		if len(line) == 0 {
			fmt.Fprintln(os.Stdout, "")
			continue
		}
		_, err := printHighlightedLine(vim, i, line, opts)
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

func getAnsiFromHl(hl map[string]any) (string, error) {
	var ansiCode strings.Builder

	if fg, ok := hl["fg"].(uint64); ok {
		if ansi := rgbToAnsi(fg); ansi != "" {
			ansiCode.WriteString(ansi)
		}
	}

	if bold, ok := hl["bold"].(bool); ok && bold == true {
		ansiCode.WriteString(AnsiBold)
	}
	if italic, ok := hl["italic"].(bool); ok && italic == true {
		ansiCode.WriteString(AnsiItalic)
	}
	if underline, ok := hl["underline"].(bool); ok && underline == true {
		ansiCode.WriteString(AnsiUnderline)
	}

	result := ansiCode.String()
	if result == "" {
		result = AnsiReset
	}
	return result, nil
}

func printHighlightedLine(vim *nvim.Nvim, lineNum int, line string, opts formatOpts) (string, error) {
	var currentAnsi string
	segments := words.NewSegmenter([]byte(line))
	col := 0
	for segments.Next() {
		var hl map[string]any
		token := segments.Text()
		token_len := len(token)
		if token_len == 1 && token[0] == '\t' {
			fmt.Fprint(os.Stdout, opts.tab)
			col += token_len
			continue
		}
		err := vim.ExecLua("return NvcatGetHl(...)", &hl, lineNum, col)
		if err != nil {
			if currentAnsi != "" {
				fmt.Fprint(os.Stderr, AnsiReset)
				currentAnsi = ""
			}
			fmt.Fprint(os.Stdout, token)
			col += token_len
			continue
		}

		ansi, err := getAnsiFromHl(hl)
		if err != nil {
			if currentAnsi != "" {
				fmt.Fprint(os.Stderr, AnsiReset)
				currentAnsi = ""
			}
			fmt.Fprint(os.Stdout, token)
			col += token_len
			continue
		}

		// Update ANSI escape sequence only if it changed
		if ansi != currentAnsi {
			if currentAnsi != "" {
				fmt.Fprint(os.Stderr, AnsiReset)
			}
			fmt.Fprint(os.Stderr, ansi)
			currentAnsi = ansi
		}

		fmt.Fprint(os.Stdout, token)
		col += token_len
	}

	// Reset color at the end of the line
	if currentAnsi != "" {
		fmt.Fprint(os.Stderr, AnsiReset)
	}

	return "", nil
}
