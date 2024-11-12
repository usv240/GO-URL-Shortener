document.addEventListener("DOMContentLoaded", function() {
    const form = document.getElementById("urlForm");

    form.addEventListener("submit", async function(event) {
        event.preventDefault(); // Prevent the form from reloading the page

        const url = document.getElementById("url").value;
        const customAlias = document.getElementById("customAlias").value;

        console.log("Submitting form with URL:", url, "and custom alias:", customAlias);

        try {
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
                displayNotification(shortCode, originalURL);
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
});

function displayNotification(shortCode, originalURL) {
    const notificationBar = document.getElementById("notificationBar");

    if (notificationBar) {
        const fakeDomain = "http://myurl.com";
        const displayURL = `${fakeDomain}/${shortCode}`;

        console.log("Displaying shortened URL:", displayURL);
        console.log("Original URL for redirection:", originalURL);

        notificationBar.textContent = `Short URL: ${displayURL}`;
        notificationBar.style.display = "block";

        // Open directly to the original URL on click
        notificationBar.onclick = function() {
            console.log("Redirecting to original URL:", originalURL);
            window.open(originalURL, "_blank"); // Open directly to the original URL in a new tab
        };
    } else {
        console.error("Notification bar element not found. Please check your HTML structure.");
    }
}
