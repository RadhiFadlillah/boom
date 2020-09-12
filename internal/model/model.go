package model

import "time"

// Metadata is metadata of a content
type Metadata struct {
	// Content related metadatas
	Title       string
	Description string
	Author      string
	CreateTime  time.Time
	UpdateTime  time.Time
	Tags        []string

	// Theme related metadatas
	Theme      string
	Template   string
	Pagination int
}
