# nvcat

A command-line utility that displays files with Neovim's syntax highlighting in the terminal.

## Overview

`nvcat` is a CLI tool similar to Unix's `cat` but with syntax highlighting powered by Neovim's syntax and treesitter engines. It leverages Neovim's capabilities to provide accurate syntax highlighting for a wide range of file formats directly in your terminal.

## Features

- Syntax highlighting using Neovim's highlighting engine
- Support for treesitter-based highlighting
- Optional line numbers
- Can use your existing Neovim configuration or run with a clean instance

## Installation

**Prequisites**:
- Neovim 0.10+ (must be accessible via `nvim`)
- A terminal that supports true color

### Prebuilt binaries

See the [releases page](https://github.com/brianhuster/nvcat/releases) for prebuilt binaries for Linux, macOS, and Windows.

### From source

Requires Go 1.22+

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

You can configure Nvcat using Vimscript or Lua just the same as you would with Neovim. However, it is recommended to start from a scratch config, because LSP, plugins can cause unnecessary long startup time and other unexpected behaviors. Generally you would only need to set colorscheme, tabstop, or enable Treesitter highlighting

Nvcat configuration should be put in `$XDG_CONFIG_HOME/nvim/init.lua` or `$XDG_CONFIG_HOME/nvim/init.vim`. Unlike Neovim configuration, Nvcat configuration is always loaded by Nvcat no matter if you use flag `-clean` or not.

On startup, Nvcat will set the variable `g:nvcat` to the current version of Nvcat. So you can also use this variable to set Nvcat-specific configurations without havin to put it in a different location than Neovim configuration.

Example:
```lua
--- ~/.config/nvim/init.lua
if not vim.g.nvcat then
    -- Load LSP, plugins, etc.
else
    vim.opt.rtp:append(path/to/your/colorscheme/runtimepath)
    -- Add runtimepath directory containing 'parser/<your-treesitter-parsers>'
    vim.opt.rtp:append("replace/with/your/actual/path")

    vim.cmd.colorscheme("your-colorscheme")
    vim.o.tabstop = 4

    vim.api.nvim_create_autocmd("BufRead", {
        callback = function()
            local ok = pcall(vim.treesitter.start)
            if not ok then
                vim.cmd.syntax("on")
            end
        end
    })
end
```

## Limitationns

- `nvcat` only supports legacy and Treesitter-based syntax highlighting engines. It does not support LSP-based highlighting.
- `nvcat` doesn't change background colors, so you should use a color scheme that has a background color similar to your terminal's

## Buy me a coffee

<a href="https://paypal.me/brianphambinhan">
    <img src="https://www.paypalobjects.com/webstatic/mktg/logo/pp_cc_mark_111x69.jpg" alt="Paypal" style="height: 69px;">
</a>
<a href="https://img.vietqr.io/image/mb-9704229209586831984-print.png?addInfo=Donate%20for%20livepreview%20plugin%20nvim&accountName=PHAM%20BINH%20AN">
    <img src="https://github.com/user-attachments/assets/f28049dc-ce7c-4975-a85e-be36612fd061" alt="VietQR" style="height: 85px;">
</a>
