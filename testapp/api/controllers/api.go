package controllers

import (
	"encoding/json"

	"testapp/api/models"
	"testapp/lib/router"
)

func RegisterAPI(api *router.RouterGroup) {

	// [READ] Get all tasks
	api.GET("/tasks", func(c *router.Context) {
		tasks := models.GetAllTasks()
		c.JSON(200, tasks)
	})

	// [CREATE] Add a new task
	api.POST("/tasks", func(c *router.Context) {
		var body struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil || body.Title == "" {
			c.JSON(400, map[string]string{"error": "Invalid request"})
			return
		}
		newTask := models.CreateTask(body.Title)
		c.JSON(201, newTask)
	})

	// [UPDATE] Toggle task status
	api.PUT("/tasks/{id}", func(c *router.Context) {
		id := c.Param("id")
		var body struct {
			Completed bool `json:"completed"`
		}
		if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
			c.JSON(400, map[string]string{"error": "Invalid body"})
			return
		}

		updatedTask, err := models.UpdateTaskStatus(id, body.Completed)
		if err != nil {
			c.JSON(404, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(200, updatedTask)
	})

	// [DELETE] Remove a task
	api.DELETE("/tasks/{id}", func(c *router.Context) {
		id := c.Param("id")
		if err := models.DeleteTask(id); err != nil {
			c.JSON(404, map[string]string{"error": err.Error()})
			return
		}
		c.JSON(200, map[string]string{"message": "Deleted successfully"})
	})
}
