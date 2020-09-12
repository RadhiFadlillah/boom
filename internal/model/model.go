package model

import "time"

// Metadata is metadata of a content
type Metadata struct {
	// Content related metadatas
	Title       string    `toml:",omitempty"`
	Description string    `toml:",omitempty"`
	Author      string    `toml:",omitempty"`
	CreateTime  time.Time `toml:",omitempty"`
	UpdateTime  time.Time `toml:",omitempty"`
	Tags        []string  `toml:",omitempty"`

	// Theme related metadatas
	Theme      string `toml:",omitempty"`
	Template   string `toml:",omitempty"`
	Pagination int    `toml:",omitempty"`
}
