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
├── assets/
└── content/
    └── _index.toml
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

This directory is used to store assets that used in the generated site. You can structure your assets freely depending on your requirements. For example :

```
.
└── assets/
    ├── image-1.jpg
    ├── image-2.png
    └── portofolio/
        └── app-1/
            ├── screenshot-1.png
            └── screenshot-2.png
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
- It's a markdown file with [valid metadata](#metadata), the simplest format and might be the one that you will use the most;
- It's `index.html` file, useful if you want to manually serve a HTML page.

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
        ├── post-3.md     // https://example.com/category-2/post-3
        └── post-4.html   // https://example.com/category-2/post-4
```

## Metadata

Every markdown file that want to be rendered must contains a valid metadata. Metadata in `boom` is stored as [TOML file][1] which fulfill struct below :

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
	Theme         string
	Template      string
	ChildTemplate string
	Pagination    int
}
```

The metadata is put above your markdown file, surrounded by `+++` (inspired by Hugo) :

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

- `title` is the title of the page. When we render the post into HTML file, this field will be put into title meta tags like `<title>`, `<meta property="og:title">` and `<meta name="twitter:title">`.
- `description` is the description of the page. This field will be put into description meta tags like `<meta name="description">`, `<meta property="og:description">` and `<meta name="twitter:description">`.
- `author` is the author of the page. This field will be put into `<meta name="author">` tag.
- `createTime` is the time when the page created. You could set it to future to make that page not build until the time is reached.
- `updateTime` is the time when the page last updated. If omitted, it will use the `createTime`.
- `tags` is the tags for the page.
- `draft` specifies whether the page is ready to publish or not. If set to `true`, this page will not be build.
- `theme` is the name of theme that will be used for the page.
- `template` is the name of template that will be used for the page.
- `childTemplate` is the name of template that will be used for the children page of current directory. Only used in `_index.md` file.
- `pagination` is the count of items for each pagination. By default it will be 10. If it sets to negative there will be no pagination.

If part of metadata is omitted, `boom` will use metadata from the page's parent directory. With that said, you should at least create `_index.md` with valid metadata in root `content` directory, as the fallback for pages with incomplete metadata.

## License

Boom is distributed under Apache-2.0 License. Basically, it means you can do what you like with the software. However, if you modify it, you have to include the license and notices, and state what did you change.

[1]: https://toml.io/en/v1.0.0-rc.1
[2]: https://gohugo.io/content-management/front-matter/