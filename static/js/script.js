document.addEventListener("DOMContentLoaded", function () {
    const form = document.getElementById("urlForm");
    const deleteButton = document.getElementById("deleteButton");
    const notificationBar = document.getElementById("notificationBar");

    // Ensure critical elements exist in the DOM
    if (!form || !deleteButton || !notificationBar) {
        console.error("Some required elements are missing in the DOM.");
        return;
    }

    // Handle form submission for shortening URLs
    form.addEventListener("submit", async function (event) {
        event.preventDefault();

        // Hide the notification bar when a new URL is submitted
        notificationBar.style.display = "none";

        const url = document.getElementById("url").value.trim();
        const customAlias = document.getElementById("customAlias").value.trim();

        console.log("Submitting form with URL:", url, "and custom alias:", customAlias);

        

        try {
            // Check if the URL or alias already exists
            const existingMapping = await checkIfURLOrAliasExists(url, customAlias);

            if (existingMapping) {
                showPopup(`A short URL for this link already exists: ${existingMapping.shortCode}`);
                return;
            }

            // Send request to create a new shortened URL
            const response = await fetch("/shorten", {
                method: "POST",
                headers: {
                    "Content-Type": "application/x-www-form-urlencoded",
                },
                body: `url=${encodeURIComponent(url)}&custom_alias=${encodeURIComponent(customAlias)}`,
            });

            if (response.ok) {
                const responseData = await response.json();
                console.log("Server response data:", responseData);

                const { shortCode, originalURL } = responseData;

                // Display the notification bar with the shortened URL
                displayNotification(shortCode, originalURL);
                showPopup(`Short URL created: ${shortCode}`);
            } else if (response.status === 409) {
                // Handle conflict error for duplicate URLs or aliases
                const errorMessage = await response.text();
                showPopup(`Conflict: ${errorMessage}`);
            } else {
                const errorMessage = await response.text();
                console.error("Error response from server:", errorMessage);
                showPopup(`Error: ${errorMessage}`);
            }
        } catch (error) {
            console.error("Error submitting form:", error);
            showPopup("An unexpected error occurred.");
        }
    });

    // Handle the Delete button click
    deleteButton.addEventListener("click", async function () {
        notificationBar.style.display = "none";

        const url = document.getElementById("url").value.trim();
        const shortCode = document.getElementById("customAlias").value.trim();

        if (!url && !shortCode) {
            showPopup("Please enter either a URL or a Shortened URL to delete.");
            return;
        }

        try {
            const response = await fetch(`/delete-url`, {
                method: "DELETE",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({ url, shortCode }),
            });

            if (response.ok) {
                const message = await response.text();
                showPopup(message);
            } else {
                const errorMessage = await response.text();
                showPopup(`Error: ${errorMessage}`);
            }
        } catch (error) {
            console.error("Error deleting URL:", error);
            showPopup("An unexpected error occurred.");
        }
    });

    async function checkIfURLOrAliasExists(url, customAlias) {
        try {
            const response = await fetch(
                `/check-url-or-alias?url=${encodeURIComponent(url)}&custom_alias=${encodeURIComponent(customAlias)}`
            );
    
            // Ensure the response status is OK and process JSON
            if (response.ok) {
                const data = await response.json(); // Parse response as JSON
                if (data.exists) {
                    return data.mapping;
                }
                return null;
            } else {
                const errorText = await response.text(); // Log error text for debugging
                console.error("Error from backend:", errorText);
                return null;
            }
        } catch (error) {
            console.error("Error checking URL or alias:", error);
            return null;
        }
    }

    function displayNotification(shortCode, originalURL) {
        const fakeDomain = "http://myurl.com";
        const displayURL = `${fakeDomain}/${shortCode}`;

        notificationBar.textContent = `Shortened URL: ${displayURL}`;
        notificationBar.style.display = "block";
        notificationBar.onclick = () => window.open(originalURL, "_blank");
    }

    function showPopup(message) {
        const popup = document.createElement("div");
        popup.className = "popup";
        popup.textContent = message;
        document.body.appendChild(popup);

        setTimeout(() => popup.remove(), 3000);
    }
});
