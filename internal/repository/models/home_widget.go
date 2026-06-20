package models

import "time"

type HomeWidget struct {
	ID              int64
	SortOrder       int32
	TagText         string
	TagBg           string
	TagColor        string
	Title           string
	Description     string
	BackgroundStyle string
	ImageURL        string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
