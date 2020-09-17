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
	Theme         string `toml:",omitempty"`
	Template      string `toml:",omitempty"`
	ChildTemplate string `toml:",omitempty"`
	Pagination    int    `toml:",omitempty"`
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
	Content     template.HTML

	// Special for non _index.md file
	Tags     []TagPath
	PrevFile ContentPath
	NextFile ContentPath
}

// ContentPath is path to a content.
type ContentPath struct {
	URLPath    string
	Title      string
	UpdateTime time.Time
	IsDir      bool
}

// TagPath is path to a content.
type TagPath struct {
	URLPath string
	Name    string
	Count   int
}
