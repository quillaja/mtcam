package main

import (
	"github.com/quillaja/mtcam/config"
)

// ScrapedConfig holds settings for the scraped executable.
type ScrapedConfig struct {
	config.SuiteConfig `json:"-"`

	ImageWidth             int
	ImageQuality           int
	ImageEqualityTesting   bool
	ImageEqualityTolerance float64

	UserAgent         string
	RequestTimeoutSec int

	GoogleTzAPIKey string

	// astro max tries?

	// wait time between attempts to schedule a day's worth of scrapes?
	// max tries for a day's scrapes
}
