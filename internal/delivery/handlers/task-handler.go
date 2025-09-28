package handlers

import (
	"betera-tz/internal/delivery/apierr"
	"betera-tz/internal/delivery/handlers/helper"
	"betera-tz/internal/domain/services"
	"betera-tz/internal/dto"
	"encoding/json"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

type TaskHandler struct {
	TaskService services.TaskService
}

func NewTaskHandler(ts services.TaskService) *TaskHandler {
	return &TaskHandler{
		TaskService: ts,
	}
}

// PostTasks godoc
// @Summary Create a new task
// @Description Create a new task with title and description
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body dto.CreateTaskRequest true "Task to create"
// @Success 201 {object} dto.CreateTaskResponse
// @Failure 400 {object} dto.ApiResponse
// @Failure 403 {object} dto.ApiResponse
// @Failure 500 {object} dto.ApiResponse
// @Router /api/v1/tasks [post]
func (th *TaskHandler) PostTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := dto.CreateTaskRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSONError(w, apierr.InvalidRequest())
		return
	}

	id, err := th.TaskService.Create(ctx, req.Title, req.Description)
	if err != nil {
		helper.WriteJSONError(w, apierr.ToApiError(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.CreateTaskResponse{
		Id: *id,
	})
}

// PatchTasksIdStatus godoc
// @Summary Update task status
// @Description Update status of a task by ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param status query dto.PatchTasksIdStatusParams true "New status"
// @Success 200 {object} dto.ApiResponse
// @Failure 400 {object} dto.ApiResponse
// @Failure 404 {object} dto.ApiResponse
// @Failure 500 {object} dto.ApiResponse
// @Router /api/v1/tasks/{id}/status [patch]
func (th *TaskHandler) PatchTasksIdStatus(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, params dto.PatchTasksIdStatusParams) {
	ctx := r.Context()
	if err := th.TaskService.UpdateStatus(ctx, id.String(), params.Status); err != nil {
		helper.WriteJSONError(w, apierr.ToApiError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dto.ApiResponse{
		Code:    http.StatusOK,
		Message: "task's status updated",
	})
}

// GetTasks godoc
// @Summary List tasks
// @Description Get a paginated list of tasks. All query parameters are optional.
// @Tags tasks
// @Accept json
// @Produce json
// @Param amount query int false "Number of tasks per page"
// @Param page query int false "Page number"
// @Param statusFilter query string false "Filter by task status" Enums(created, processing, done)
// @Success 200 {array} dto.TaskResponse "List of tasks"
// @Failure 400 {object} dto.ApiResponse "Bad request"
// @Failure 500 {object} dto.ApiResponse "Internal server error"
// @Router /api/v1/tasks [get]
func (th *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request, params dto.GetTasksParams) {
	ctx := r.Context()

	var (
		amount       int
		page         int
		statusFilter string
	)
	if params.Amount == nil || params.Page == nil || *params.Amount <= 0 || *params.Page <= 0 {
		amount = -1
		page = 1
	} else {
		amount = *params.Amount
		page = *params.Page
	}

	def := ""

	if params.StatusFilter == nil {
		statusFilter = def
	} else {
		statusFilter = string(*params.StatusFilter)
	}

	tasks, err := th.TaskService.Get(ctx, amount, page, statusFilter)
	if err != nil {
		helper.WriteJSONError(w, apierr.ToApiError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

// GetTasksId godoc
// @Summary Get task by ID
// @Description Get detailed information about a task by its ID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ApiResponse
// @Failure 404 {object} dto.ApiResponse
// @Failure 500 {object} dto.ApiResponse
// @Router /api/v1/tasks/{id} [get]
func (th *TaskHandler) GetTasksId(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ctx := r.Context()
	task, err := th.TaskService.GetById(ctx, id.String())
	if err != nil {
		helper.WriteJSONError(w, apierr.ToApiError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}
