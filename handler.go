package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

// generateShortCode creates a random short code
func generateShortCode(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// indexHandler serves the main page with the form
func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

// shortenURLHandler processes URL shortening requests
func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	longURL := r.FormValue("url")
	customAlias := r.FormValue("custom_alias")

	// Ensure the URL has a protocol (http:// or https://)
	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		longURL = "http://" + longURL
	}

	// Step 1: Check if the original URL already has a short URL
	existingMapping, err := getURLMappingByOriginalURL(longURL)
	if err != nil {
		http.Error(w, "Error checking existing URL", http.StatusInternalServerError)
		return
	}
	if existingMapping != nil {
		// If a custom alias is requested but the original URL already has a short URL
		if customAlias != "" && existingMapping.ShortCode != customAlias {
			http.Error(w, "A short URL for this link already exists", http.StatusConflict)
			return
		}

		// Return existing short code and original URL in JSON
		responseData := map[string]string{
			"shortCode":   existingMapping.ShortCode,
			"originalURL": existingMapping.OriginalURL,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responseData)
		return
	}

	// Step 2: If a custom alias is provided, check if it already exists
	if customAlias != "" {
		aliasMapping, err := getURLMapping(customAlias)
		if err != nil {
			http.Error(w, "Error checking custom alias", http.StatusInternalServerError)
			return
		}
		if aliasMapping != nil {
			http.Error(w, "Custom alias already in use", http.StatusConflict)
			return
		}
	}

	// Step 3: Generate a short code if no custom alias is provided
	shortCode := customAlias
	if shortCode == "" {
		shortCode = generateShortCode(8)
	}

	// Step 4: Save the new URL mapping
	err = saveURLMapping(URLMapping{
		ShortCode:      shortCode,
		OriginalURL:    longURL,
		CreatedAt:      time.Now(),
		ExpirationDate: time.Now().AddDate(0, 0, 7),
	})
	if err != nil {
		http.Error(w, "Error saving URL mapping", http.StatusInternalServerError)
		return
	}

	// Respond with JSON containing shortCode and originalURL
	responseData := map[string]string{
		"shortCode":   shortCode,
		"originalURL": longURL,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseData)
}

// redirectHandler redirects to the original URL using the short code
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/r/")

	// Retrieve the original URL based on the shortcode
	longURL, exists := getURLFromCache(shortCode)
	if !exists {
		mapping, err := getURLMapping(shortCode)
		if err != nil || mapping == nil {
			http.NotFound(w, r)
			return
		}
		longURL = mapping.OriginalURL
		cacheURL(shortCode, longURL)
	}

	// Ensure the URL has a protocol (http:// or https://)
	if !strings.HasPrefix(longURL, "http") {
		longURL = "http://" + longURL
	}

	// Redirect directly to the original URL
	http.Redirect(w, r, longURL, http.StatusFound)
}

// renderTemplate renders HTML templates
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
