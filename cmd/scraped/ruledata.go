package main

import (
	"time"

	"github.com/quillaja/mtcam/astro"
	"github.com/quillaja/mtcam/model"
)

// RulesData is the data passed to the template when a camera's "rules" are evaluated.
type RulesData struct {
	Astro    astro.Data
	Now      time.Time
	Mountain model.Mountain
	Camera   model.Camera
}

// UrlData is the data passed to the template when a camera's "url" is evaluated.
type UrlData struct {
	Now      time.Time
	Mountain model.Mountain
	Camera   model.Camera
}
