# nvcat

A command-line utility that displays files with Neovim's syntax highlighting in the terminal.

## Overview

`nvcat` is a CLI tool similar to Unix's `cat` but with syntax highlighting powered by Neovim's syntax and treesitter engines. It leverages Neovim's capabilities to provide accurate syntax highlighting for a wide range of file formats directly in your terminal.

## Features

- Syntax highlighting using Neovim's highlighting engine
- Support for treesitter-based highlighting
- Optional line numbers
- Can use your existing Neovim configuration or run with a clean instance

<table width="100%">
  <tr>
    <td width="600px">
      <img src="https://github.com/user-attachments/assets/cfc76a64-e0da-4ac1-8860-cf14a091e9fc" />
    </td>
  </tr>
  <tr>
    <th>Viewing a Lua file with <a href="https://github.com/folke/tokyonight.nvim">Tokyonight-night colorscheme</a></th>
  </tr>
  <tr>
    <td width="600px">
      <img src="https://github.com/user-attachments/assets/79d53df0-6fcc-4ada-a36d-42179e447bd3" />
    </td>      
  </tr>
  <tr>
    <th>
      Nvcat as <a href="https://github.com/junegunn/fzf">fzf</a>'s preview, with my custom 
      <a href="https://github.com/brianhuster/dotfiles/blob/5e7aed6/nvim/colors/an.lua">colorscheme</a>.
    </th>
  </tr>
</table>

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

### Options

* `-n`: Show line numbers
* `-clean`: Don't load Neovim's config
* `-v`: Show version information
* `-h`: Show help information

## Configuration

You can configure Nvcat using Vimscript or Lua just the same as you would with Neovim. However, it is recommended to start from a scratch config, because LSP, plugins can cause unnecessary long startup time and other unexpected behaviors. Generally you would only need to set colorscheme, tabstop, or enable Treesitter highlighting

There are 2 ways to configure Nvcat:

#### 1. Use Nvcat's config directory: `$XDG_CONFIG_HOME/nvcat/init.lua` or `$XDG_CONFIG_HOME/nvcat/init.vim`.

With this method, your Nvcat configuration will be seperated from your Neovim configuration, and it can be loaded even when the flag `-clean` is given

Example:
```lua
--- ~/.config/nvcat/init.lua
vim.opt.rtp:append(path/to/your/colorscheme/runtimepath)
-- Add runtimepath directory containing 'parser/<your-treesitter-parsers>'
vim.opt.rtp:append("replace/with/your/actual/path")

vim.cmd.colorscheme("your-colorscheme")
vim.o.tabstop = 4

vim.api.nvim_create_autocmd("FileType", {
    callback = function()
        pcall(vim.treesitter.start)
    end
})
```

#### 2. Use Neovim's config dictionary (`:echo stdpath('config')`).

Nvcat sets Vimscript variable `g:nvcat` on startup, so you can use it to control which parts of your Neovim configuration should not be used by Nvcat.

Example:
```lua
--- ~/.config/nvim/init.lua
if vim.g.nvcat then
    -- Nvcat configuration
else
    -- LSP, plugins, etc.
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

## Credits

- [neovim/go-client](https://github.com/neovim/go-client) : Controlling Nvim from Go
- [clipperhouse/uax29](https://github.com/clipperhouse/uax29): Unicode graphemes support
