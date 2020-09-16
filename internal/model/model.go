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
	IsDir      bool
	URLPath    string
	ActiveTag  string
	PathTrails []PagePath

	Title       string
	Description string
	Author      string
	CreateTime  time.Time
	UpdateTime  time.Time
	Tags        []PagePath
	Content     template.HTML
	Pagination  int

	theme    string
	template string
}

// PagePath is path to the page and title of each directory.
type PagePath struct {
	Path  string
	Title string
}
