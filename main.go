package main

import (
	"flag"
	"fmt"
	"github.com/neovim/go-client/nvim"
	"os"
	"path/filepath"
	"strings"
)

const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	dim       = "\033[2m"
	italic    = "\033[3m"
	underline = "\033[4m"
)

var (
	lineNumbers = flag.Bool("n", false, "Show line numbers")
	theme       = flag.String("theme", "default", "Neovim colorscheme to use")
)

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
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

	vim, err := nvim.NewChildProcess(nvim.ChildProcessArgs("-u", "NONE", "--embed", "--headless"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting Neovim: %v\n", err)
		os.Exit(1)
	}
	defer vim.Close()

	vim.Command(fmt.Sprintf("colorscheme %s", *theme))
	err = vim.Command(fmt.Sprintf("edit %s", absFilename))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}

	var buffer nvim.Buffer
	buffer, err = vim.CurrentBuffer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current buffer: %v\n", err)
		os.Exit(1)
	}

	fileContent, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(string(fileContent), "\n")

	var filetype string
	vim.BufferOption(buffer, "filetype", &filetype)

	fmt.Printf("%s%s%s (%s)\n", bold, filename, reset, filetype)

	fmt.Println(strings.Repeat("─", 40))

	processFile(vim, lines)
}

func processFile(vim *nvim.Nvim, lines []string) {
	err := loadHighlightDefinitions(vim)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not load highlight definitions: %v\n", err)
	}
	for i, line := range lines {
		linePrefix := ""
		if *lineNumbers {
			linePrefix = fmt.Sprintf("%s%4d %s", dim, i+1, reset)
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
			return vim.api.nvim_get_hl(0, { id = hl_id, link = false })
		end
		local hl_id = captures[#captures].id
		return vim.api.nvim_get_hl(0, { id = hl_id, link = false })
	end
	`
	return vim.ExecLua(script, nil, nil)
}

func rgbToAnsi(color int) string {
	r := uint8((color >> 16) & 0xFF)
	g := uint8((color >> 8) & 0xFF)
	b := uint8(color & 0xFF)
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

func getHighlightColor(vim *nvim.Nvim, hl map[string]any) (string, error) {
	var ansiCode strings.Builder

	if fg, ok := hl["fg"].(string); ok && fg != "" {
		var hexColor int
		hexColor, err := vim.ColorByName(fg)
		if err == nil {
			if ansi := rgbToAnsi(hexColor); ansi != "" {
				ansiCode.WriteString(ansi)
			}
		}
	}

	if bold, ok := hl["bold"].(string); ok && bold == "1" {
		ansiCode.WriteString(bold)
	}
	if italic, ok := hl["italic"].(string); ok && italic == "1" {
		ansiCode.WriteString(italic)
	}
	if underline, ok := hl["underline"].(string); ok && underline == "1" {
		ansiCode.WriteString(underline)
	}

	result := ansiCode.String()
	if result == "" {
		result = reset
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
				highlightedLine.WriteString(reset)
				currentAnsi = ""
			}
			highlightedLine.WriteByte(line[col])
			continue
		}

		ansi, err := getHighlightColor(vim, hl)
		if err != nil {
			if currentAnsi != "" {
				highlightedLine.WriteString(reset)
				currentAnsi = ""
			}
			highlightedLine.WriteByte(line[col])
			continue
		}

		// Update ANSI escape sequence only if it changed
		if ansi != currentAnsi {
			if currentAnsi != "" {
				highlightedLine.WriteString(reset)
			}
			highlightedLine.WriteString(ansi)
			currentAnsi = ansi
		}

		highlightedLine.WriteByte(line[col])
	}

	// Reset color at the end of the line
	if currentAnsi != "" {
		highlightedLine.WriteString(reset)
	}

	return highlightedLine.String(), nil
}
