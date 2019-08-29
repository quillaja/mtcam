package main

import (
	"time"

	"github.com/quillaja/mtcam/astro"
	"github.com/quillaja/mtcam/model"
)

type RulesData struct {
	Astro    astro.Data
	Now      time.Time
	Mountain model.Mountain
	Camera   model.Camera
}

type UrlData struct {
	Now      time.Time
	Mountain model.Mountain
	Camera   model.Camera
}
