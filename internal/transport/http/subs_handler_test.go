// internal/transport/http/subs_handler_test.go
package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/klyakssa/aggregation-sub/internal/domain/subs"
	"github.com/klyakssa/aggregation-sub/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockService реализация мока для Service интерфейса
type MockService struct {
	mock.Mock
}

func (m *MockService) CreateSubscription(ctx context.Context, sub subs.Subscription) (subs.Subscription, error) {
	args := m.Called(ctx, sub)
	return args.Get(0).(subs.Subscription), args.Error(1)
}

func (m *MockService) GetSubscription(ctx context.Context, id int) (subs.Subscription, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(subs.Subscription), args.Error(1)
}

func (m *MockService) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]subs.Subscription, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]subs.Subscription), args.Error(1)
}

func (m *MockService) UpdateSubscription(ctx context.Context, sub subs.Subscription) (subs.Subscription, error) {
	args := m.Called(ctx, sub)
	return args.Get(0).(subs.Subscription), args.Error(1)
}

func (m *MockService) DeleteSubscription(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) ListSubscriptions(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]subs.Subscription, int, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(3)
	}
	return args.Get(0).([]subs.Subscription), args.Int(1), args.Error(3)
}

func (m *MockService) GetTotalCostByPeriod(ctx context.Context, cost subs.CostCalculation) (int, error) {
	args := m.Called(ctx, cost)
	return args.Int(0), args.Error(1)
}

func createTestContext(method, path string, body *bytes.Buffer) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, body)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	c.Request = req

	return c, w
}

func TestSubscriptionHandler_CreateSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "успешное создание подписки",
			requestBody: subs.CreateSub{
				ServiceName: "Netflix",
				Price:       799,
				UserID:      uuid.New(),
				StartDate:   subs.MonthYear{Time: time.Date(2023, 1, 0, 0, 0, 0, 0, time.UTC)},
				EndDate:     nil,
			},
			mockSetup: func(m *MockService) {
				m.On("CreateSubscription", mock.Anything, mock.AnythingOfType("subs.Subscription")).
					Return(subs.Subscription{ID: 1}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
		{
			name: "успешное создание подписки с датой окончания",
			requestBody: subs.CreateSub{
				ServiceName: "Spotify",
				Price:       199,
				UserID:      uuid.New(),
				StartDate:   subs.MonthYear{Time: time.Date(2024, 1, 0, 0, 0, 0, 0, time.UTC)},
				EndDate:     &[]subs.MonthYear{{Time: time.Date(2024, 12, 0, 0, 0, 0, 0, time.UTC)}}[0],
			},
			mockSetup: func(m *MockService) {
				m.On("CreateSubscription", mock.Anything, mock.AnythingOfType("subs.Subscription")).
					Return(subs.Subscription{ID: 2}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
		{
			name:           "ошибка binding - неверный JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   nil,
		},
		{
			name: "ошибка валидации - end_date раньше start_date",
			requestBody: subs.CreateSub{
				ServiceName: "Netflix",
				Price:       799,
				UserID:      uuid.New(),
				StartDate:   subs.MonthYear{Time: time.Date(2024, 12, 0, 0, 0, 0, 0, time.UTC)},
				EndDate:     &[]subs.MonthYear{{Time: time.Date(2024, 1, 0, 0, 0, 0, 0, time.UTC)}}[0],
			},
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "End date cannot be before start date",
			},
		},
		{
			name: "ошибка сервиса - ErrFailedToCreateSubscription",
			requestBody: subs.CreateSub{
				ServiceName: "Netflix",
				Price:       799,
				UserID:      uuid.New(),
				StartDate:   subs.MonthYear{Time: time.Date(2024, 1, 0, 0, 0, 0, 0, time.UTC)},
				EndDate:     nil,
			},
			mockSetup: func(m *MockService) {
				m.On("CreateSubscription", mock.Anything, mock.AnythingOfType("subs.Subscription")).
					Return(subs.Subscription{}, subs.ErrFailedToCreateSubscription)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to create subscription",
			},
		},
		{
			name: "ошибка сервиса - другая ошибка",
			requestBody: subs.CreateSub{
				ServiceName: "Netflix",
				Price:       799,
				UserID:      uuid.New(),
				StartDate:   subs.MonthYear{Time: time.Date(2024, 1, 0, 0, 0, 0, 0, time.UTC)},
				EndDate:     nil,
			},
			mockSetup: func(m *MockService) {
				m.On("CreateSubscription", mock.Anything, mock.AnythingOfType("subs.Subscription")).
					Return(subs.Subscription{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			tt.mockSetup(mockService)

			testLogger := &logger.Logger{Logger: zap.NewNop()}
			handler := NewSubscriptionHandler(testLogger, mockService)

			var reqBody []byte
			var err error

			switch v := tt.requestBody.(type) {
			case string:
				reqBody = []byte(v)
			case subs.CreateSub:
				reqBody, err = json.Marshal(v)
				assert.NoError(t, err)
			default:
				if tt.requestBody != nil {
					reqBody, err = json.Marshal(v)
					assert.NoError(t, err)
				}
			}

			c, w := createTestContext(http.MethodPost, "/api/subscriptions", bytes.NewBuffer(reqBody))
			handler.CreateSubscription(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				for key, value := range tt.expectedBody {
					assert.Equal(t, value, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestSubscriptionHandler_GetSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		subscriptionID string
		mockSetup      func(*MockService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "успешное получение подписки",
			subscriptionID: "1",
			mockSetup: func(m *MockService) {
				expectedSub := subs.Subscription{
					ID:          1,
					ServiceName: "Netflix",
					Price:       799,
				}
				m.On("GetSubscription", mock.Anything, 1).Return(expectedSub, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: subs.Subscription{
				ID:          1,
				ServiceName: "Netflix",
				Price:       799,
			},
		},
		{
			name:           "неверный ID",
			subscriptionID: "invalid",
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid subscription ID",
			},
		},
		{
			name:           "подписка не найдена",
			subscriptionID: "999",
			mockSetup: func(m *MockService) {
				m.On("GetSubscription", mock.Anything, 999).Return(subs.Subscription{}, subs.ErrFailedToFoundSubscription)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Subscription not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			tt.mockSetup(mockService)

			testLogger := &logger.Logger{Logger: zap.NewNop()}
			handler := NewSubscriptionHandler(testLogger, mockService)

			c, w := createTestContext(http.MethodGet, "/api/subscriptions/"+tt.subscriptionID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.subscriptionID}}
			handler.GetSubscription(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				if expectedSub, ok := tt.expectedBody.(subs.Subscription); ok {
					var response subs.Subscription
					json.Unmarshal(w.Body.Bytes(), &response)
					assert.Equal(t, expectedSub.ID, response.ID)
					assert.Equal(t, expectedSub.ServiceName, response.ServiceName)
				} else if expectedMap, ok := tt.expectedBody.(map[string]interface{}); ok {
					var response map[string]interface{}
					json.Unmarshal(w.Body.Bytes(), &response)
					for key, value := range expectedMap {
						assert.Equal(t, value, response[key])
					}
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestSubscriptionHandler_DeleteSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		subscriptionID string
		mockSetup      func(*MockService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "успешное удаление подписки",
			subscriptionID: "1",
			mockSetup: func(m *MockService) {
				m.On("DeleteSubscription", mock.Anything, 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   nil,
		},
		{
			name:           "неверный ID - буквы вместо цифр",
			subscriptionID: "abc",
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid subscription ID",
			},
		},
		{
			name:           "неверный ID - отрицательное число",
			subscriptionID: "-1",
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid subscription ID",
			},
		},
		{
			name:           "неверный ID - слишком большое число",
			subscriptionID: "99999999999999999999",
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid subscription ID",
			},
		},
		{
			name:           "подписка не найдена",
			subscriptionID: "999",
			mockSetup: func(m *MockService) {
				m.On("DeleteSubscription", mock.Anything, 999).Return(subs.ErrFailedToFoundSubscription)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "Subscription not found",
			},
		},
		{
			name:           "ошибка удаления подписки",
			subscriptionID: "1",
			mockSetup: func(m *MockService) {
				m.On("DeleteSubscription", mock.Anything, 1).Return(subs.ErrFailedToDeleteSubscription)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "failed to delete subscription",
			},
		},
		{
			name:           "ошибка получения количества затронутых строк",
			subscriptionID: "1",
			mockSetup: func(m *MockService) {
				m.On("DeleteSubscription", mock.Anything, 1).Return(subs.ErrFailedToGetRowsAffected)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "failed to get rows affected",
			},
		},
		{
			name:           "общая ошибка сервера",
			subscriptionID: "1",
			mockSetup: func(m *MockService) {
				m.On("DeleteSubscription", mock.Anything, 1).Return(errors.New("database connection error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil,
		},
		{
			name:           "пустой ID",
			subscriptionID: "",
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid subscription ID",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			tt.mockSetup(mockService)

			testLogger := &logger.Logger{Logger: zap.NewNop()}
			handler := NewSubscriptionHandler(testLogger, mockService)

			c, w := createTestContext(http.MethodDelete, "/api/subscriptions/"+tt.subscriptionID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.subscriptionID}}

			handler.DeleteSubscription(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				for key, value := range tt.expectedBody {
					assert.Equal(t, value, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestSubscriptionHandler_ListSubscriptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockService)
		expectedStatus int
	}{
		{
			name:        "успешный список подписок",
			queryParams: "?user_id=" + userID.String() + "&page=1&page_size=10",
			mockSetup: func(m *MockService) {
				subscriptions := []subs.Subscription{
					{ID: 1, UserID: userID, ServiceName: "Netflix"},
					{ID: 2, UserID: userID, ServiceName: "Spotify"},
				}
				m.On("ListSubscriptions", mock.Anything, userID, 1, 10).
					Return(subscriptions, 2, 1, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "отсутствует user_id",
			queryParams:    "?page=1&page_size=10",
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "неверный формат user_id",
			queryParams:    "?user_id=invalid&page=1&page_size=10",
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			tt.mockSetup(mockService)

			testLogger := &logger.Logger{Logger: zap.NewNop()}
			handler := NewSubscriptionHandler(testLogger, mockService)

			c, w := createTestContext(http.MethodGet, "/api/subscriptions"+tt.queryParams, nil)
			handler.ListSubscriptions(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestSubscriptionHandler_GetTotalCost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockService)
		expectedStatus int
	}{
		{
			name:        "успешный расчет стоимости",
			queryParams: "?start_date=2024-01&end_date=2024-12",
			mockSetup: func(m *MockService) {
				m.On("GetTotalCostByPeriod", mock.Anything, mock.AnythingOfType("subs.CostCalculation")).
					Return(5000, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "отсутствует start_date",
			queryParams:    "?end_date=2024-12-31",
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "неверный формат start_date",
			queryParams:    "?start_date=01-2024&end_date=2024-12",
			mockSetup:      func(m *MockService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockService)
			tt.mockSetup(mockService)

			testLogger := &logger.Logger{Logger: zap.NewNop()}
			handler := NewSubscriptionHandler(testLogger, mockService)

			c, w := createTestContext(http.MethodGet, "/api/subscriptions/total-cost"+tt.queryParams, nil)
			handler.GetTotalCost(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
