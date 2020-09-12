# Boom

Boom is a simple static site generator for creating a simple site. "Boom" apparently means "tree" in Dutch language, so I thought it's a good name because :

- It's easy to say;
- It's suitable for the app purpose since the content of site will be structured like a tree inside directories.

## Features

- Easy to install and use;
- Doesn't require any external dependency;
- Easy and flexible to theme.

## CLI Usage

```
Simple static site generator

Usage:
  boom [command]

Available Commands:
  help        Help about any command
  new         Create a new site at specified path
  server      Run webserver for the site

Flags:
  -h, --help   help for boom

Additional help topics:
  boom build  Build the static site

Use "boom [command] --help" for more information about a command.
```

## Directory Structure

Running `boom new .` from the command line will create a directory with the following elements :

```
.
├── themes/
└── content/
    ├── _index.md
    └── _meta.toml
```

### "themes" directory

This directory is used to store themes that used in the generated site. You can store several themes with each theme separated in their respected directory :

```
.
└── themes/
    ├── theme-1
    ├── theme-2
    └── theme-N
```

### "content" directory

This directory is used to store all content of the site. In the most basic blog, it will looks like this :

```
.
└── content/
    ├── post-1/
    ├── post-2/
    └── post-N/
```

In `boom`, a directory is considered as a page (and will be rendered) if it fulfills at least one of these criterias :

- It contains `index.html` file;
- It contains `_index.md` and optionally a valid `_meta.toml` file;
- It contains sub directories which fulfill one of two criterias above.

So, if we expand the tree from previous content structure, it might looks like this :

```
.
└── content/
    ├── _index.md
    ├── _meta.toml
    ├── post-1/
    │   ├── _index.md
    │   └── _meta.toml
    ├── post-2/
    │   ├── _index.md
    │   └── _meta.toml
    └── post-3/
        └── index.html
```

As long as you follow three rules above, you can customize your content structure to follow your requirements. For example, here is structure for website with several categories :

```
.
└── content/
    ├── _index.md
    ├── _meta.toml
    ├── category-1/
    │   ├── post-1/
    │   │   ├── _index.md
    │   │   └── _meta.toml
    │   └── post-2/
    │       ├── _index.md
    │       └── _meta.toml
    └── category-2/
        └── post-3/
            └── _index.html
```

## Metadata File

In a directory that uses `_index.md` as its entry, you will need to provide metadata file named `_meta.toml`. As its name imply, metadata in `boom` is stored as [TOML file][1] which fulfill struct below :

```go
type Metadata struct {
	// Content related metadatas
	Title       string
	Description string
	Author      string
	CreateTime  time.Time
	UpdateTime  time.Time
	Tags        []string
	Draft       bool

	// Theme related metadatas
	Theme      string
	Template   string
	Pagination int
}
```

Here are the explanation for each field :

- `title` is the title of the page. When we render the post into HTML file, this field will be put into title meta tags like `<title>`, `<meta property="og:title">` and `<meta name="twitter:title">`.
- `description` is the description of the page. This field will be put into description meta tags like `<meta name="description">`, `<meta property="og:description">` and `<meta name="twitter:description">`.
- `author` is the author of the page. This field will be put into `<meta name="author">` tag.
- `createTime` is the time when the page created. You could set it to future to make that page not build until the time is reached.
- `updateTime` is the time when the page last updated. If omitted, it will use the `createTime`.
- `tags` is the tags for the page.
- `draft` specifies whether the page is ready to publish or not. If set to `true`, this page will not be build.
- `theme` is the name of theme that will be used for the page.
- `template` is the name of template that will be used for the page.
- `pagination` is the count of items for each pagination. By default it will be 10. If it sets to negative there will be no pagination.

If `_meta.toml` file is omitted, `boom` will use metadata from the page's parent directory. With that said, `_meta.toml` is optional except for root `content` directory.

## License

Boom is distributed under Apache-2.0 License. Basically, it means you can do what you like with the software. However, if you modify it, you have to include the license and notices, and state what did you change.

[1]: https://toml.io/en/v1.0.0-rc.1