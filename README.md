# GO URL Shortener

A URL shortener built with Go (Golang) and MongoDB, designed to convert long URLs into short, easy-to-remember links. This project also supports custom aliases, automatic expiration, and efficient caching.

## Features
- **Short URL Generation**: Quickly generate shortened URLs for any long URL.
- **Custom Aliases**: Allow users to create custom short URLs (e.g., `myshort.ly/customalias`).
- **Automatic Expiration**: URLs automatically expire after 7 days, with background cleanup.
- **Caching**: Cache popular redirects for fast access and reduced database load.
- **User-Friendly Interface**: A clean and simple web interface with instant notifications.

## Tech Stack
- **Backend**: Go (Golang)
- **Database**: MongoDB for persistent storage
- **Frontend**: HTML, CSS, and JavaScript (for interactivity)

Getting Started
Prerequisites
    Golang (Go 1.17+)
    MongoDB (running on localhost:27017 by default)

Installation
1. Clone the repository:

    git clone https://github.com/usv240/GO-URL-Shortener.git
    cd GO-URL-Shortener

2. Install dependencies: Make sure MongoDB is running on your machine.

3. Run the application:

        go run main.go storage.go handler.go cache.go background.go

4. Access the application: Open your browser and go to http://localhost:8080 to access the URL Shortener interface.

Environment Configuration (Optional)

You can customize the MongoDB connection string and other settings by modifying the initializeDatabase() function in storage.go.

**Usage**

1. Shorten a URL:
        Enter a long URL in the input field.
        Optionally, provide a custom alias for the short URL.
        Click Shorten URL to generate a short URL.

2. Redirect using Shortened URL:
        The generated short URL will be displayed at the bottom.
        Clicking on the shortened URL will open the original link in a new tab.

3. API Endpoints:
        POST /shorten: Submit a URL with an optional custom alias to get a shortened URL.
        GET /r/{shortCode}: Redirects to the original URL.

**API Documentation**
**POST /shorten**

  Description: Creates a shortened URL.
  Parameters:
      url (required): The original URL to shorten.
      custom_alias (optional): A custom alias for the shortened URL.
  Response: JSON object with shortCode and originalURL.

GET /r/{shortCode}

  Description: Redirects to the original URL corresponding to the provided short code.
  Response: 302 Redirect to the original URL.

Example Request & Response

Request:

    curl -X POST -d "url=https://www.google.com" -d "custom_alias=example" http://localhost:8080/shorten

Response:

    {
      "shortCode": "example",
      "originalURL": "https://www.google.com"
    }

**Advanced Features**
  Automatic Expiration: URLs expire after 7 days and are removed by a background cleanup process.
  Caching Layer: Frequently accessed URLs are cached to reduce database load and improve performance.
  Custom Alias Validation: Ensures custom aliases are unique and not already taken.



Future Improvements
  User Authentication: Allow users to manage their URLs.
  Analytics: Track the number of clicks per shortened URL.
  Custom Expiration Times: Let users set custom expiration dates for each URL.

Contributing: Contributions are welcome! Please open an issue or submit a pull request if you'd like to improve the project.

Contact: For questions or support, please reach out to ujwalv098@gmail.com
