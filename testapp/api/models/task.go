package models

import (
	"fmt"
	"math/rand"
	"time"
)

type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

var tasksDB = []Task{
	{ID: "t1", Title: "Learn Spidey Framework", Completed: true},
	{ID: "t2", Title: "Build an interactive CRUD app", Completed: false},
	{ID: "t3", Title: "Master Vanilla JS", Completed: false},
}

func GetAllTasks() []Task {
	return tasksDB
}

func CreateTask(title string) Task {
	// Seed to avoid deterministic IDs
	rand.Seed(time.Now().UnixNano())
	newTask := Task{
		ID:        fmt.Sprintf("t%d", rand.Intn(100000)),
		Title:     title,
		Completed: false,
	}
	tasksDB = append(tasksDB, newTask)
	return newTask
}

func UpdateTaskStatus(id string, completed bool) (Task, error) {
	for i, t := range tasksDB {
		if t.ID == id {
			tasksDB[i].Completed = completed
			return tasksDB[i], nil
		}
	}
	return Task{}, fmt.Errorf("task not found")
}

func DeleteTask(id string) error {
	for i, t := range tasksDB {
		if t.ID == id {
			tasksDB = append(tasksDB[:i], tasksDB[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task not found")
}
