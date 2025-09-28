package client

import (
	"betera-tz/internal/domain/models"
	"betera-tz/internal/dto"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

func Run() {
	c, err := NewClient("http://localhost:3333")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	log.Printf("creating task...")

	createReq := dto.PostTasksJSONRequestBody{
		Title:       toPtr("Test task9"),
		Description: toPtr("Test description8"),
	}

	resp, err := c.PostTasks(ctx, createReq)
	if err != nil {
		log.Printf("Failed to create task: %v", err)
		return
	}
	defer resp.Body.Close()

	var taskResp dto.CreateTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		log.Printf("Failed to decode task: %v", err)
		return
	}

	if resp.StatusCode == 201 {
		fmt.Printf("Task created successfully: %s\n", taskResp.Id.String())
	} else {
		fmt.Printf("Failed to create task: %d\n", resp.StatusCode)
	}

	log.Printf("fetching task...")

	params := &dto.GetTasksParams{
		Page:   toPtr(1),
		Amount: toPtr(10),
	}

	tasksResp, err := c.GetTasks(ctx, params)
	if err != nil {
		log.Printf("Faield to fetch tasks: %v", err)
		return
	}
	defer tasksResp.Body.Close()

	if tasksResp.StatusCode == 200 {
		var tasks []models.Task
		if err := json.NewDecoder(tasksResp.Body).Decode(&tasks); err != nil {
			log.Printf("Failed to decode tasks list: %v", err)
			return
		}

		fmt.Println("Tasks fetched successfully:")
		for _, t := range tasks {
			fmt.Printf("ID: %s, Title: %s, Status: %s\n", t.ID.String(), t.Title, t.Status)
		}
	} else {
		fmt.Printf("Failed to fetch tasks: %d\n", tasksResp.StatusCode)
	}

	log.Printf("updating task's status...")

	statusParams := &dto.PatchTasksIdStatusParams{
		Status: "created",
	}

	statusResp, err := c.PatchTasksIdStatus(ctx, taskResp.Id, statusParams)
	if err != nil {
		log.Printf("Failed to update status: %v", err)
		return
	}
	defer statusResp.Body.Close()

	if statusResp.StatusCode == 200 {
		fmt.Println("Status updated successfully")
	} else {
		fmt.Printf("Failed to update status: %d\n", statusResp.StatusCode)
	}

	log.Println("receiving task by id...")

	task, err := c.GetTasksId(ctx, taskResp.Id)
	if err != nil {
		log.Printf("Failed to receive task by id: %v", err)
		return
	}
	defer task.Body.Close()

	taskInfo2 := &models.Task{}

	if err := json.NewDecoder(task.Body).Decode(taskInfo2); err != nil {
		log.Printf("Failed to decode task: %v", err)
		return
	}

	if statusResp.StatusCode == 200 {
		log.Printf("Task received by id successfully: %v", taskInfo2)
	} else {
		fmt.Printf("Failed to receive task by id: %d\n", statusResp.StatusCode)
	}
}

func toPtr[T any](v T) *T {
	return &v
}
