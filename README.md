# nook

`nook` is a minimal and fast static site generator built with Go.

It automatically mirrors the structure of your local Markdown files,
and outputs unstyled HTML files.

## Quick Start

Simply run `nook` to build your local directory into a static website, outputted
to the `public` directory by default.

Run `nook help` for more options.

This command will clean the output directory (changed to `dist`), as well as
copy markdown files (great for LLMs):

```bash
nook . --clean --copy-markdown --out-directory dist
```

## Styling

If `css` files are found in the input directory, they are automatically included
into all of the pages (added in lexical order).

Sample themes can be found in `/themes`.
