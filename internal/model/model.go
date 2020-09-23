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
	Theme            string `toml:",omitempty"`
	DirTemplate      string `toml:",omitempty"`
	FileTemplate     string `toml:",omitempty"`
	TagFilesTemplate string `toml:",omitempty"`
	Pagination       int    `toml:",omitempty"`
}

// DirData is data that used when rendering a directory.
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

// FileData is data that used when rendering a file.
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

// TagFilesData is template model for rendering a tag file list.
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
