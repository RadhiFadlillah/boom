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
  build       Build the static site
  help        Help about any command
  new         Create a new site or metadata
  server      Run webserver for the site

Flags:
  -h, --help   help for boom

Use "boom [command] --help" for more information about a command.
```

## Directory Structure

Running `boom new .` from the command line will create a directory with the following elements :

```
.
├── themes/
├── assets/
└── content/
    └── _index.md
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

### "assets" directory

This directory is used to store assets that used in the generated site. You can structure your assets freely depending on your requirements. Just remember that the structure will be used later for URL path. For example :

```
.
└── assets/
    ├── image-1.jpg					// https://example.com/assets/image-1.jpg
    ├── image-2.png					// https://example.com/assets/image-2.png
    └── portofolio/
        └── app-1/
            ├── screenshot-1.png	// https://example.com/assets/portofolio/app-1/screenshot-1.png
            └── screenshot-2.png	// https://example.com/assets/portofolio/app-1/screenshot-2.png
```

### "content" directory

This directory is used to store all content of the site. In the most basic blog, it will looks like this :

```
.
└── content/
    ├── _index.md
    ├── post-1.md
    ├── post-2.md
    ├── post-3.md
    └── post-4.md
```

In `boom`, a file or directory is considered as a page (and will be rendered) if it fulfills at least one of these criterias :

- It's a directory which contains `_index.md` file;
- It's a markdown file with `.md` extension, the simplest format and might be the one that you will use the most.

As long as you follow rules above, you can customize your content structure to follow your requirements. The structure will later be used for page URL. For example, here is structure for website with several categories :

```
.
└── content/
    ├── _index.md         // https://example.com
    ├── category-1/
    │   ├── _index.md     // https://example.com/category-1
    │   ├── post-1.md     // https://example.com/category-1/post-1
    │   └── post-2.md     // https://example.com/category-1/post-2
    └── category-2/
        ├── _index.md     // https://example.com/category-2
        └── post-3.md     // https://example.com/category-2/post-3
```

The only limitations for content structure are :

- There must be no directories named `assets` and `themes` in root of `content` directory;
- There must be no files named `assets.md` and `themes.md` in root of `content` directory;
- There must be no files named `tag-*.md` and directories named `tag-*` anywhere within content directory.

## Metadata

It's recommended to put metadata on your markdown file. Metadata in `boom` is stored as [TOML file][1] which fulfill struct below :

```go
type Metadata struct {
	// Content's metadatas
	Title       string    `toml:",omitempty"`
	Description string    `toml:",omitempty"`
	Author      string    `toml:",omitempty"`
	CreateTime  time.Time `toml:",omitempty"`
	UpdateTime  time.Time `toml:",omitempty"`
	Tags        []string  `toml:",omitempty"`
	Draft       bool      `toml:",omitempty"`

	// Theme's metadatas
	Theme            string `toml:",omitempty"`
	DirTemplate      string `toml:",omitempty"`
	FileTemplate     string `toml:",omitempty"`
	TagFilesTemplate string `toml:",omitempty"`
	Pagination       int    `toml:",omitempty"`
}
```

The metadata is put above your markdown file, surrounded by `+++` (inspired by [Hugo][2]) :

```
+++
Author = "Radhi Fadlillah"
CreateTime = 2020-09-12T15:24:00+07:00
Pagination = 10
Title = "Learn Boom"
+++

Your markdown content
```

Here are the explanation for each field :

- `Title` is the title of the page. When we render the post into HTML file, this field will be put into title meta tags like `<title>`, `<meta property="og:title">` and `<meta name="twitter:title">`.
- `Description` is the description of the page. This field will be put into description meta tags like `<meta name="description">`, `<meta property="og:description">` and `<meta name="twitter:description">`.
- `Author` is the author of the page. This field will be put into `<meta name="author">` tag.
- `CreateTime` is the time when the page created. You could set it to future to make that page not build until the time is reached.
- `UpdateTime` is the time when the page last updated. If omitted, it will use the `createTime`.
- `Tags` is the tags for the page.
- `Draft` specifies whether the page is ready to publish or not. If set to `true`, this page will not be build.
- `Theme` is the name of theme that will be used for the page.
- `DirTemplate` is the name for template that will be used for rendering current and child directory. Default is `directory`.
- `FileTemplate` is the name for template that will be used for rendering current file or files inside current directory. Default is `file`.
- `TagFilesTemplate` is the name for template that will be used for rendering list of files for each tag in current directory. Default is `tagfiles`.
- `Pagination` is the count of items for each pagination. If it sets to less or equal zero there will be no pagination.

If part of metadata is omitted, `boom` will use metadata from the page's parent directory. With that said, you must at least create `_index.md` with valid metadata in root `content` directory, as the fallback for pages with incomplete metadata.

## Theming

Say we want to create a theme called `simple` for our site. To do so, we need to create a directory with following structure :

```
.
└── themes/
    └── simple/
        ├── directory.html
        ├── file.html
        └── tagfiles.html
```

Those three HTML will be later used for rendering HTML files for our site :

- `directory.html` is template for rendering directory;
- `file.html` is template for rendering `*.md` files;
- `tagfiles.html` is template for rendering list of files with specified tag.

The rendering process will uses [`html/template`][3] package that provided by Go standard library. There are three structs that used as data when rendering templates :

- `DirData` is data that used when rendering directory :

	```go
	type DirData struct {
		URLPath    string
		PathTrails []ContentPath

		Title       string
		Description string
		Author      string
		Content     template.HTML
		ChildItems  []ContentPath
		ChildTags   []TagPath

		PageSize    int
		CurrentPage int
		MaxPage     int
	}
	```

- `FileData` is data that used when rendering file :

	```go
	type FileData struct {
		URLPath    string
		PathTrails []ContentPath

		Title       string
		Description string
		Author      string
		CreateTime  time.Time
		UpdateTime  time.Time
		Content     template.HTML

		Tags     []TagPath
		PrevFile ContentPath
		NextFile ContentPath
	}
	```

- `TagFilesData` is data that used when rendering `tagfiles` template :

	```go
	type TagFilesData struct {
		URLPath    string
		PathTrails []ContentPath
		ActiveTag  string

		Title       string
		Files       []ContentPath
		PageSize    int
		CurrentPage int
		MaxPage     int
	}
	```

As you can see, all of those data structs use `ContentPath` and `TagPath` which structured like this :

```go
// ContentPath is path to a content.
type ContentPath struct {
	// Common
	IsDir   bool
	URLPath string
	Title   string

	// File only
	UpdateTime time.Time

	// Dir only
	NChild int
}

// TagPath is path to a tag files.
type TagPath struct {
	URLPath string
	Name    string
	Count   int
}
```

## License

Boom is distributed under Apache-2.0 License. Basically, it means you can do what you like with the software. However, if you modify it, you have to include the license and notices, and state what did you change.

[1]: https://toml.io/en/v1.0.0-rc.1
[2]: https://gohugo.io/content-management/front-matter/
[3]: https://golang.org/pkg/html/template/