package main

import (
	"github.com/quillaja/mtcam/config"
)

// ScrapedConfig holds settings for the scraped executable.
type ScrapedConfig struct {
	config.SuiteConfig `json:"-"`

	UserAgent         string
	RequestTimeoutSec int

	GoogleTzAPIKey string

	Image Image

	Scheduling Scheduling

	// astro max tries?
}

// Image holds settings related to processing scraped images.
type Image struct {
	Width             int
	Quality           int
	EqualityTesting   bool
	EqualityTolerance float64
}

// Scheduling holds settings related to scheduling tasks.
type Scheduling struct {
	// max tries for a day's scrapes
	MaxAttempts int
	// wait time between attempts to schedule a day's worth of scrapes?
	WaitTime int //mins?
}
