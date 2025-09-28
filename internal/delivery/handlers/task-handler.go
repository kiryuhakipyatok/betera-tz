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

func (th *TaskHandler) PostTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := dto.CreateTaskRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.WriteJSONError(w, apierr.InvalidRequest())
		return
	}

	id, err := th.TaskService.Create(ctx, *req.Title, *req.Description)
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

func (th *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request, params dto.GetTasksParams) {
	ctx := r.Context()
	var (
		amount int
		page   int
	)
	if params.Amount == nil || params.Page == nil {
		amount = -1
		page = 0
	} else {
		amount = *params.Amount
		page = *params.Page
	}
	tasks, err := th.TaskService.Get(ctx, amount, page)
	if err != nil {
		helper.WriteJSONError(w, apierr.ToApiError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

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
