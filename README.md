# nvcat

A command-line utility that displays files with Neovim's syntax highlighting in the terminal.

## Overview

`nvcat` (En-vee-cat) is a CLI tool similar to Unix's `cat` but with syntax highlighting powered by Neovim's syntax and treesitter engines. It leverages Neovim's capabilities to provide accurate syntax highlighting for a wide range of file formats directly in your terminal.

## Features

- Syntax highlighting using Neovim's highlighting engine
- Support for treesitter-based highlighting
- Optional line numbers
- Can use your existing Neovim configuration or run with a clean instance

## Installation

- Prequisites: Neovim 0.10+ (must be accessible via `nvim`)

### Prebuilt binaries

See the [releases page](https://github.com/brianhuster/nvcat/releases) for prebuilt binaries for Linux, macOS, and Windows.

### From source

```bash
go install github.com/brianhuster/nvimcat@latest
```

Or clone and build manually:

```bash
git clone https://github.com/brianhuster/nvcat.git
cd nvcat
sudo make install
```

## Usage

```bash
nvcat [options] <file>
```

Run `nvcat -h` for more information.

## Configuration

Nvcat configuration is basically the same as Neovim's configuration, you can put it in `$XDG_CONFIG_HOME/nvim/init.lua` or `$XDG_CONFIG_HOME/nvim/init.vim`. Unlike Neovim configuration, Nvcat configuration is always loaded by Nvcat no matter if you use flag `-clean` or not.

## Limitations

- `nvcat` only supports legacy and Treesitter-based syntax highlighting engines. It does not support LSP-based highlighting.
- `nvcat` doesn't change background colors, so you should use a color scheme that has a background color similar to your terminal's

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.
