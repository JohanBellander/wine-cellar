# Wine Cellar Inventory App

![CI](https://github.com/JohanBellander/wine-cellar/actions/workflows/ci.yml/badge.svg)

This is a Wine Cellar Inventory application built with Go.

## Prerequisites

- Go 1.18 or higher

## Running the Application

1.  Open a terminal in the project root.
2.  Run the following command:

    ```bash
    go run main.go
    ```

3.  Open your browser and navigate to [http://localhost:8080](http://localhost:8080).

## Features

-   **List Wines**: View your wine inventory.
-   **Add Wine**: Add new wines to your collection.
-   **Wine Details**: View detailed information about a specific wine.

## Project Structure

-   `main.go`: The main application logic and HTTP server.
-   `templates/`: HTML templates for the UI.
    -   `list.html`: The main dashboard/list view.
    -   `wine_form.html`: The form to add or edit a wine.
    -   `details.html`: The detailed view of a wine.
-   `mockups/`: Original HTML mockups provided.

## Notes

-   The backend is currently mocked with in-memory storage. Data will be lost when the server restarts.
-   Images for new wines are currently placeholders.
