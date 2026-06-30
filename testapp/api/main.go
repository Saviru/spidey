package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"test/lib/pages"
	"test/lib/router"
)

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Issue struct {
	ID        string `json:"id"`
	ProjectID string `json:"projectId"`
	Title     string `json:"title"`
	Status    string `json:"status"` // "todo", "in_progress", "done"
}

var (
	projects = make(map[string]Project)
	issues   = make(map[string]Issue)
	dataMu   sync.RWMutex
	nextID   = 1
)

func generateID() string {
	id := nextID
	nextID++
	return fmt.Sprintf("%d", id)
}

// ----- PROJECTS API -----

func GetProjects(c *router.Context) {
	dataMu.RLock()
	defer dataMu.RUnlock()
	list := make([]Project, 0, len(projects))
	for _, p := range projects {
		list = append(list, p)
	}
	c.JSON(http.StatusOK, list)
}

func CreateProject(c *router.Context) {
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	dataMu.Lock()
	defer dataMu.Unlock()
	id := generateID()
	p := Project{ID: id, Name: input.Name}
	projects[id] = p
	c.JSON(http.StatusCreated, p)
}

func DeleteProject(c *router.Context) {
	id := c.Param("id")
	dataMu.Lock()
	defer dataMu.Unlock()
	delete(projects, id)
	// Delete associated issues
	for issueID, issue := range issues {
		if issue.ProjectID == id {
			delete(issues, issueID)
		}
	}
	c.JSON(http.StatusOK, map[string]bool{"success": true})
}

// ----- ISSUES API -----

func GetProjectIssues(c *router.Context) {
	projectID := c.Param("id")
	dataMu.RLock()
	defer dataMu.RUnlock()
	list := make([]Issue, 0)
	for _, i := range issues {
		if i.ProjectID == projectID {
			list = append(list, i)
		}
	}
	c.JSON(http.StatusOK, list)
}

func CreateIssue(c *router.Context) {
	var input struct {
		ProjectID string `json:"projectId"`
		Title     string `json:"title"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	dataMu.Lock()
	defer dataMu.Unlock()
	id := generateID()
	i := Issue{
		ID:        id,
		ProjectID: input.ProjectID,
		Title:     input.Title,
		Status:    "todo",
	}
	issues[id] = i
	c.JSON(http.StatusCreated, i)
}

func UpdateIssue(c *router.Context) {
	id := c.Param("id")
	var input struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&input); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	dataMu.Lock()
	defer dataMu.Unlock()
	i, exists := issues[id]
	if !exists {
		c.JSON(http.StatusNotFound, map[string]string{"error": "Issue not found"})
		return
	}
	i.Status = input.Status
	issues[id] = i
	c.JSON(http.StatusOK, i)
}

func DeleteIssue(c *router.Context) {
	id := c.Param("id")
	dataMu.Lock()
	defer dataMu.Unlock()
	delete(issues, id)
	c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func main() {
	app := router.New()
	app.Static("/assets/", "public/assets")

	api := app.Group("/api")
	api.GET("/projects", GetProjects)
	api.POST("/projects", CreateProject)
	api.DELETE("/projects/{id}", DeleteProject)

	api.GET("/projects/{id}/issues", GetProjectIssues)
	api.POST("/issues", CreateIssue)
	api.PUT("/issues/{id}", UpdateIssue)
	api.DELETE("/issues/{id}", DeleteIssue)

	pages.RegisterRoutes(app)

	// Since we had the port conflict earlier, let's use 8081 consistently
	fmt.Println("Server listening on http://localhost:8081")
	if err := app.Listen("8081"); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
