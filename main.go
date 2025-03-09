package main

import (
	"flag"
	"fmt"
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

var (
	lineNumbers = flag.Bool("n", false, "Show line numbers")
	clean       = flag.Bool("clean", false, "Use a clean Neovim instance")
	help        = flag.Bool("h", false, "Show help")
	tab = ""
)

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 || *help {
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
	if *clean {
		args = append(args, "--clean")
	}
	vim, err := nvim.NewChildProcess(nvim.ChildProcessArgs(args...))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Neovim: %v\n", err)
		os.Exit(1)
	}
	defer vim.Close()

	vim.RegisterHandler("redraw", func(args []any) {})
	vim.AttachUI(2 * len(lines), 80, map[string]any{})

	err = vim.ExecLua(`
	local joinpath = vim.fs.joinpath
	local config_dir = joinpath(vim.fn.fnamemodify(vim.fn.stdpath('config'), ':h'), 'nvcat')
	vim.opt.rtp:append(config_dir)
	if vim.fn.filereadable(joinpath(config_dir, 'init.lua')) == 1 then
		vim.cmd.source(joinpath(config_dir, 'init.lua'))
		return
	end
	if vim.fn.filereadable(joinpath(config_dir, 'init.vim')) == 1 then
		vim.cmd.source(joinpath(config_dir, 'init.vim'))
	end
	`, nil, nil)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
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
	tab = strings.Repeat(" ", tabstop)

	fmt.Printf("%s%s%s\n", Bold, filename, Reset)
	fmt.Println(strings.Repeat("─", 40))

	err = vim.Command(fmt.Sprintf("edit %s", absFilename))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}

	processFile(vim, lines)
}

func processFile(vim *nvim.Nvim, lines []string) {
	err := loadHighlightDefinitions(vim)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load highlight definitions: %v\n", err)
	}
	numDigits := len(fmt.Sprintf("%d", len(lines)))
	for i, line := range lines {
		linePrefix := ""
		if *lineNumbers {
			format := "%s%" + strconv.Itoa(numDigits) + "d %s"
			linePrefix = fmt.Sprintf(format, Dim, i+1, Reset)
		}
		if len(line) == 0 {
			fmt.Println(linePrefix)
			continue
		}
		highlightedLine, err := getHighlightedLine(vim, i, line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting highlights for line %d: %v\n", i+1, err)
			fmt.Println(linePrefix + line)
			continue
		}
		highlightedLine = strings.ReplaceAll(highlightedLine, "\t", tab)
		fmt.Println(linePrefix + highlightedLine)
	}
}

func loadHighlightDefinitions(vim *nvim.Nvim) error {
	script := `
	function GetHl(row, col)
		local captures = vim.treesitter.get_captures_at_pos(0, row, col)
		if #captures == 0 then
			local hl_id = vim.fn.synID(row + 1, col + 1, 1)
			if hl_id == 0 then
				return vim.empty_dict()
			end
			return vim.api.nvim_get_hl(0, { id = hl_id, link = false, create = false })
		end
		local hl_name = '@' .. captures[#captures].capture
		return vim.api.nvim_get_hl(0, { name = hl_name, link = false, create = false })
	end
	`
	return vim.ExecLua(script, nil, nil)
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

func getHighlightedLine(vim *nvim.Nvim, lineNum int, line string) (string, error) {
	var highlightedLine strings.Builder
	var currentAnsi string

	for col := range len(line) {
		var hl map[string]any
		err := vim.ExecLua("return GetHl(...)", &hl, lineNum, col)
		if err != nil {
			if currentAnsi != "" {
				highlightedLine.WriteString(Reset)
				currentAnsi = ""
			}
			highlightedLine.WriteByte(line[col])
			continue
		}

		ansi, err := getHighlightColor(hl)
		if err != nil {
			if currentAnsi != "" {
				highlightedLine.WriteString(Reset)
				currentAnsi = ""
			}
			highlightedLine.WriteByte(line[col])
			continue
		}

		// Update ANSI escape sequence only if it changed
		if ansi != currentAnsi {
			if currentAnsi != "" {
				highlightedLine.WriteString(Reset)
			}
			highlightedLine.WriteString(ansi)
			currentAnsi = ansi
		}

		highlightedLine.WriteByte(line[col])
	}

	// Reset color at the end of the line
	if currentAnsi != "" {
		highlightedLine.WriteString(Reset)
	}

	return highlightedLine.String(), nil
}
