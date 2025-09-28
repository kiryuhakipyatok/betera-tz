package client

import (
	"betera-tz/internal/domain/models"
	"betera-tz/internal/dto"
	"context"
	"encoding/json"
	"fmt"
	"log"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type TestClient struct {
	Client *Client
}

func CreateClient() *TestClient {
	c, err := NewClient("http://localhost:3333/api/v1")
	if err != nil {
		panic(err)
	}

	return &TestClient{
		Client: c,
	}
}

func (c *TestClient) CreateNewTask(title, description string) *openapi_types.UUID {
	log.Printf("creating task...")
	createReq := dto.PostTasksJSONRequestBody{
		Title:       title,
		Description: description,
	}
	ctx := context.Background()
	resp, err := c.Client.PostTasks(ctx, createReq)
	if err != nil {
		log.Println(fmt.Errorf("failed to create task: %w", err))
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		log.Println(fmt.Errorf("failed to create task: %d", resp.StatusCode))
		return nil
	}

	var taskResp dto.CreateTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&taskResp); err != nil {
		log.Println(fmt.Errorf("failed to decode CreateTaskResponse: %w", err))
		return nil
	}

	log.Printf("Task created successfully: %s\n", taskResp.Id.String())
	return &taskResp.Id
}

func (c *TestClient) FetchTasks(amount, page *int, status *string) {
	log.Printf("fetching task...")

	ctx := context.Background()

	var filter *dto.GetTasksParamsStatusFilter
	if status != nil {
		tmp := dto.GetTasksParamsStatusFilter(*status)
		filter = &tmp
	}

	req := &dto.GetTasksParams{
		Page:         page,
		Amount:       amount,
		StatusFilter: filter,
	}

	tasksResp, err := c.Client.GetTasks(ctx, req)
	if err != nil {
		log.Printf("Failed to fetch tasks: %v", err)
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

}

func (c *TestClient) UpdateTaskStatus(id openapi_types.UUID, status string) {
	log.Printf("updating task's status...")

	ctx := context.Background()

	statusParams := &dto.PatchTasksIdStatusParams{
		Status: status,
	}

	statusResp, err := c.Client.PatchTasksIdStatus(ctx, id, statusParams)
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
}

func (c *TestClient) GetTaskById(id openapi_types.UUID) {
	log.Println("receiving task by id...")

	ctx := context.Background()

	taskResp, err := c.Client.GetTasksId(ctx, id)
	if err != nil {
		log.Printf("Failed to receive task by id: %v", err)
		return
	}
	defer taskResp.Body.Close()

	taskInfo2 := &models.Task{}

	if err := json.NewDecoder(taskResp.Body).Decode(taskInfo2); err != nil {
		log.Printf("Failed to decode task: %v", err)
		return
	}

	if taskResp.StatusCode == 200 {
		log.Printf("Task received by id successfully: %v", taskInfo2)
	} else {
		fmt.Printf("Failed to receive task by id: %d\n", taskResp.StatusCode)
	}
}
