package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

// generateShortCode creates a random short code
func generateShortCode(length int) string {
	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Println("Error generating random short code:", err)
		return ""
	}
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

	// Normalize the URL (prepend "http://" if missing)
	if strings.HasPrefix(longURL, "www.") {
		longURL = "http://" + longURL
	} else if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
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
			log.Println("Short code not found:", shortCode)
			http.NotFound(w, r)
			return
		}
		longURL = mapping.OriginalURL
		cacheURL(shortCode, longURL)
	}

	// Redirect directly to the original URL
	http.Redirect(w, r, longURL, http.StatusFound)
}

// checkURLOrAliasHandler checks if a given URL or custom alias exists
func checkURLOrAliasHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	customAlias := r.URL.Query().Get("custom_alias")

	// Ensure at least one parameter is provided
	if url == "" && customAlias == "" {
		http.Error(w, "URL or Custom Alias must be provided", http.StatusBadRequest)
		return
	}

	// Prepare the filter for MongoDB query
	filter := bson.M{}
	if url != "" {
		filter["original_url"] = url
	}
	if customAlias != "" {
		filter["short_code"] = customAlias
	}

	// Query MongoDB for the mapping
	var existingMapping URLMapping
	err := urlCollection.FindOne(context.Background(), filter).Decode(&existingMapping)
	if err == mongo.ErrNoDocuments {
		json.NewEncoder(w).Encode(map[string]interface{}{"exists": false})
		return
	} else if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}

	// Match found, return the mapping
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exists":  true,
		"mapping": existingMapping,
	})
}

// renderTemplate renders HTML templates
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// deleteURLHandler deletes a URL mapping by its short code or original URL
func deleteURLHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		URL       string `json:"url"`
		ShortCode string `json:"shortCode"`
	}

	// Parse JSON body
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Ensure at least one field is provided
	if requestData.URL == "" && requestData.ShortCode == "" {
		http.Error(w, "Either URL or Short Code must be provided", http.StatusBadRequest)
		return
	}

	var filter bson.M

	// Normalize and set filter
	if requestData.URL != "" {
		normalizedURL := requestData.URL
		if !strings.HasPrefix(normalizedURL, "http://") && !strings.HasPrefix(normalizedURL, "https://") {
			normalizedURL = "http://" + normalizedURL
		}
		filter = bson.M{"original_url": normalizedURL}
	} else if requestData.ShortCode != "" {
		filter = bson.M{"short_code": requestData.ShortCode}
	}

	// Check if the URL or Short Code exists
	var existingMapping URLMapping
	err = urlCollection.FindOne(context.Background(), filter).Decode(&existingMapping)
	if err == mongo.ErrNoDocuments {
		http.Error(w, "No matching URL or Short Code found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error querying database", http.StatusInternalServerError)
		return
	}

	// Delete the URL
	result, err := urlCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, "Error deleting URL", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Failed to delete URL", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "URL deleted successfully.")
}
