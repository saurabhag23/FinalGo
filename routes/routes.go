package routes

import (
    "encoding/json"
    "io"
    "log"
    "net/http"

    "github.com/gofiber/fiber/v2"
)

type Post struct {
    UserID int    `json:"userId"`
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Body   string `json:"body"`
}

func GetPosts(c *fiber.Ctx) error {
    id := c.Params("id") // Get the post ID from the request URL parameters
    if id == "" {
        log.Fatal("Invalid ID")
    }

    // Fetch the post data from the API
    resp, err := http.Get("https://jsonplaceholder.typicode.com/posts/" + id)
    if err != nil {
        return c.JSON(fiber.Map{"error": err.Error()})
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return c.JSON(fiber.Map{"error": err.Error()})
    }

    var data Post
    err = json.Unmarshal(body, &data)
    if err != nil {
        return c.JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(data)

}
