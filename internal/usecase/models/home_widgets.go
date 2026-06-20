package models

import (
	"errors"
	"strings"
)

var ErrHomeWidgetNotFound = errors.New("ErrHomeWidgetNotFound")

type HomeWidgetItem struct {
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
}

type ListHomeWidgetsOutput struct {
	Data []HomeWidgetItem
}

type AdminCreateHomeWidgetInput struct {
	SortOrder       int32
	TagText         string
	TagBg           string
	TagColor        string
	Title           string
	Description     string
	BackgroundStyle string
	ImageURL        string
	IsActive        bool
}

func (i *AdminCreateHomeWidgetInput) Normalize() {
	i.TagText = strings.TrimSpace(i.TagText)
	i.TagBg = strings.TrimSpace(i.TagBg)
	i.TagColor = strings.TrimSpace(i.TagColor)
	i.Title = strings.TrimSpace(i.Title)
	i.Description = strings.TrimSpace(i.Description)
	i.BackgroundStyle = strings.TrimSpace(i.BackgroundStyle)
	i.ImageURL = strings.TrimSpace(i.ImageURL)
	if i.TagBg == "" {
		i.TagBg = "rgba(60,200,100,0.85)"
	}
	if i.TagColor == "" {
		i.TagColor = "#06291a"
	}
	if i.BackgroundStyle == "" && i.ImageURL == "" {
		i.BackgroundStyle = "linear-gradient(135deg,#1a1030,#2a1840)"
	}
}

type AdminUpdateHomeWidgetInput struct {
	ID              int64
	SortOrder       *int32
	TagText         *string
	TagBg           *string
	TagColor        *string
	Title           *string
	Description     *string
	BackgroundStyle *string
	ImageURL        *string
	IsActive        *bool
}

type AdminHomeWidgetOutput struct {
	HomeWidgetItem
}
