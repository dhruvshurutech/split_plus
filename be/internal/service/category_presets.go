package service

import (
	"regexp"
	"strings"
)

// CategoryPreset represents a system category template
type CategoryPreset struct {
	Slug  string
	Name  string
	Icon  string
	Color string
}

// SystemCategoryPresets is the list of predefined category templates
var SystemCategoryPresets = []CategoryPreset{
	{Slug: "food-and-drink", Name: "Food & Drink", Icon: "🍔", Color: "#FF6B6B"},
	{Slug: "entertainment", Name: "Entertainment", Icon: "🎬", Color: "#4ECDC4"},
	{Slug: "home", Name: "Home", Icon: "🏠", Color: "#45B7D1"},
	{Slug: "transportation", Name: "Transportation", Icon: "🚗", Color: "#96CEB4"},
	{Slug: "shopping", Name: "Shopping", Icon: "🛍️", Color: "#FFEAA7"},
	{Slug: "utilities", Name: "Utilities", Icon: "💡", Color: "#DFE6E9"},
	{Slug: "healthcare", Name: "Healthcare", Icon: "⚕️", Color: "#74B9FF"},
	{Slug: "travel", Name: "Travel", Icon: "✈️", Color: "#A29BFE"},
	{Slug: "other", Name: "Other", Icon: "📌", Color: "#B2BEC3"},
}

// GenerateSlug creates a URL-friendly slug from a name
func GenerateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (keep only alphanumeric and hyphens)
	slug = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(slug, "")
	// Remove consecutive hyphens
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	return slug
}

// GetPresetBySlug finds a preset by its slug
func GetPresetBySlug(slug string) *CategoryPreset {
	for i := range SystemCategoryPresets {
		if SystemCategoryPresets[i].Slug == slug {
			return &SystemCategoryPresets[i]
		}
	}
	return nil
}
