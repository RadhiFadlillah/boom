package model

import (
	"html/template"
	"time"
)

// Metadata is metadata of a content.
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
	Theme           string `toml:",omitempty"`
	Template        string `toml:",omitempty"`
	ChildTemplate   string `toml:",omitempty"`
	TagListTemplate string `toml:",omitempty"`
	Pagination      int    `toml:",omitempty"`
}

// PageTemplate is template model for rendering a page.
type PageTemplate struct {
	URLPath    string
	PathTrails []ContentPath

	// Dir data
	DirTitle    string
	DirItems    []ContentPath
	DirTags     []TagPath
	PageSize    int
	CurrentPage int
	MaxPage     int

	// File data
	Title       string
	Description string
	Author      string
	CreateTime  time.Time
	UpdateTime  time.Time
	Content     template.HTML `json:"-"`

	// Special for non _index.md file
	Tags     []TagPath
	PrevFile ContentPath
	NextFile ContentPath
}

// TagListTemplate is template model for rendering a tag list.
type TagListTemplate struct {
	URLPath    string
	PathTrails []ContentPath
	ActiveTag  string

	// Dir data
	DirTitle    string
	Files       []ContentPath
	PageSize    int
	CurrentPage int
	MaxPage     int
}

// ContentPath is path to a content.
type ContentPath struct {
	IsDir      bool
	URLPath    string
	Title      string
	UpdateTime time.Time
}

// TagPath is path to a content.
type TagPath struct {
	URLPath string
	Name    string
	Count   int
}
