package services

import (
	"betera-tz/internal/config"
	"betera-tz/internal/domain/models"
	"betera-tz/pkg/errs"
	"betera-tz/pkg/logger"
	"betera-tz/pkg/queue"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *models.Task) (*string, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*string), args.Error(1)
}

func (m *MockTaskRepository) GetById(ctx context.Context, id string) (*models.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) Get(ctx context.Context, amount, page int, statusFilter string) ([]models.Task, error) {
	args := m.Called(ctx, amount, page, statusFilter)
	return args.Get(0).([]models.Task), args.Error(1)
}

func (m *MockTaskRepository) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendMessage(message queue.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func TestTaskService_Create(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		description    string
		mockSetup      func(*MockTaskRepository, *MockProducer)
		expectedError  bool
		expectedResult *uuid.UUID
	}{
		{
			name:        "successful task creation",
			title:       "Test Task",
			description: "Test Description",
			mockSetup: func(mockRepo *MockTaskRepository, mockProducer *MockProducer) {
				taskId := "550e8400-e29b-41d4-a716-446655440000"
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(&taskId, nil)
				mockProducer.On("SendMessage", mock.AnythingOfType("queue.Message")).Return(nil)
			},
			expectedError:  false,
			expectedResult: func() *uuid.UUID { id, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000"); return &id }(),
		},
		{
			name:        "repository error",
			title:       "Test Task",
			description: "Test Description",
			mockSetup: func(mockRepo *MockTaskRepository, mockProducer *MockProducer) {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(nil, errors.New("database error"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
		{
			name:        "producer error but task created",
			title:       "Test Task",
			description: "Test Description",
			mockSetup: func(mockRepo *MockTaskRepository, mockProducer *MockProducer) {
				taskId := "550e8400-e29b-41d4-a716-446655440000"
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Task")).Return(&taskId, nil)
				mockProducer.On("SendMessage", mock.AnythingOfType("queue.Message")).Return(errors.New("kafka error"))
			},
			expectedError:  false,
			expectedResult: func() *uuid.UUID { id, _ := uuid.Parse("550e8400-e29b-41d4-a716-446655440000"); return &id }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockRepo := new(MockTaskRepository)
			mockProducer := new(MockProducer)
			logger := logger.NewLogger(config.AppConfig{Name: "test", Version: "1.0.0", Env: "test", LogPath: "test.log"}, "app")

			tt.mockSetup(mockRepo, mockProducer)

			service := &taskService{
				TaskRepository: mockRepo,
				Producer:       mockProducer,
				Logger:         logger,
			}

			result, err := service.Create(context.Background(), tt.title, tt.description)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.expectedResult != nil {
					assert.Equal(t, *tt.expectedResult, *result)
				}
			}

			mockRepo.AssertExpectations(t)
			mockProducer.AssertExpectations(t)
		})
	}
}

func TestTaskService_GetById(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockSetup      func(*MockTaskRepository)
		expectedError  bool
		expectedResult *models.Task
	}{
		{
			name: "successful get by id",
			id:   "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func(mockRepo *MockTaskRepository) {
				task := &models.Task{
					ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
					Title:       "Test Task",
					Description: "Test Description",
					Status:      "created",
				}
				mockRepo.On("GetById", mock.Anything, "550e8400-e29b-41d4-a716-446655440000").Return(task, nil)
			},
			expectedError: false,
			expectedResult: &models.Task{
				ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				Title:       "Test Task",
				Description: "Test Description",
				Status:      "created",
			},
		},
		{
			name: "task not found",
			id:   "550e8400-e29b-41d4-a716-446655440000",
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("GetById", mock.Anything, "550e8400-e29b-41d4-a716-446655440000").Return(nil, errs.ErrNotFound("test"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			logger := logger.NewLogger(config.AppConfig{Name: "test", Version: "1.0.0", Env: "test", LogPath: "test.log"}, "app")

			tt.mockSetup(mockRepo)

			service := &taskService{
				TaskRepository: mockRepo,
				Logger:         logger,
			}

			result, err := service.GetById(context.Background(), tt.id)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_Get(t *testing.T) {
	tests := []struct {
		name           string
		amount         int
		page           int
		statusFilter   string
		mockSetup      func(*MockTaskRepository)
		expectedError  bool
		expectedResult []models.Task
	}{
		{
			name:   "successful get with pagination",
			amount: 10,
			page:   1,
			mockSetup: func(mockRepo *MockTaskRepository) {
				tasks := []models.Task{
					{
						ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
						Title:       "Task 1",
						Description: "Description 1",
						Status:      "created",
					},
					{
						ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
						Title:       "Task 2",
						Description: "Description 2",
						Status:      "done",
					},
				}
				mockRepo.On("Get", mock.Anything, 10, 1).Return(tasks, nil)
			},
			expectedError: false,
			expectedResult: []models.Task{
				{
					ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
					Title:       "Task 1",
					Description: "Description 1",
					Status:      "created",
				},
				{
					ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
					Title:       "Task 2",
					Description: "Description 2",
					Status:      "done",
				},
			},
		},
		{
			name:   "get all tasks (no pagination)",
			amount: 0,
			page:   0,
			mockSetup: func(mockRepo *MockTaskRepository) {
				tasks := []models.Task{
					{
						ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
						Title:       "Task 1",
						Description: "Description 1",
						Status:      "created",
					},
				}
				mockRepo.On("Get", mock.Anything, 0, 0).Return(tasks, nil)
			},
			expectedError: false,
			expectedResult: []models.Task{
				{
					ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
					Title:       "Task 1",
					Description: "Description 1",
					Status:      "created",
				},
			},
		},
		{
			name:   "repository error",
			amount: 10,
			page:   1,
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("Get", mock.Anything, 10, 1).Return([]models.Task{}, errors.New("database error"))
			},
			expectedError:  true,
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			logger := logger.NewLogger(config.AppConfig{Name: "test", Version: "1.0.0", Env: "test", LogPath: "test.log"}, "app")

			tt.mockSetup(mockRepo)

			service := &taskService{
				TaskRepository: mockRepo,
				Logger:         logger,
			}

			result, err := service.Get(context.Background(), tt.amount, tt.page, tt.statusFilter)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_UpdateStatus(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		status        string
		mockSetup     func(*MockTaskRepository)
		expectedError bool
	}{
		{
			name:   "successful status update",
			id:     "550e8400-e29b-41d4-a716-446655440000",
			status: "processing",
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("UpdateStatus", mock.Anything, "550e8400-e29b-41d4-a716-446655440000", "processing").Return(nil)
			},
			expectedError: false,
		},
		{
			name:   "task not found",
			id:     "550e8400-e29b-41d4-a716-446655440000",
			status: "processing",
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("UpdateStatus", mock.Anything, "550e8400-e29b-41d4-a716-446655440000", "processing").Return(errs.ErrNotFound("test"))
			},
			expectedError: true,
		},
		{
			name:   "repository error",
			id:     "550e8400-e29b-41d4-a716-446655440000",
			status: "processing",
			mockSetup: func(mockRepo *MockTaskRepository) {
				mockRepo.On("UpdateStatus", mock.Anything, "550e8400-e29b-41d4-a716-446655440000", "processing").Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			logger := logger.NewLogger(config.AppConfig{Name: "test", Version: "1.0.0", Env: "test", LogPath: "test.log"}, "app")

			tt.mockSetup(mockRepo)

			service := &taskService{
				TaskRepository: mockRepo,
				Logger:         logger,
			}

			err := service.UpdateStatus(context.Background(), tt.id, tt.status)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
