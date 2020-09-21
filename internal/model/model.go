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

// DirTemplate is template model for rendering a directory.
type DirTemplate struct {
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

// FileTemplate is template model for rendering a file.
type FileTemplate struct {
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

// TagFilesTemplate is template model for rendering a tag file list.
type TagFilesTemplate struct {
	URLPath    string
	PathTrails []ContentPath
	ActiveTag  string

	Title       string
	Files       []ContentPath
	PageSize    int
	CurrentPage int
	MaxPage     int
}

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

// TagPath is path to a content.
type TagPath struct {
	URLPath string
	Name    string
	Count   int
}
