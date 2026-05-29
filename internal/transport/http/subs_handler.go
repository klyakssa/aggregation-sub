package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/klyakssa/aggregation-sub/internal/domain/subs"
	"github.com/klyakssa/aggregation-sub/internal/logger"
	"go.uber.org/zap"
)

type SubscriptionHandler struct {
	log     *logger.Logger
	service subs.Service
}

func NewSubscriptionHandler(log *logger.Logger, service subs.Service) *SubscriptionHandler {
	return &SubscriptionHandler{
		log:     log,
		service: service,
	}
}

// CreateSubscription создает новую подписку
//
//	@Summary		Создание подписки
//	@Description	Создает новую подписку для пользователя. Дата должна быть в формате MM-YYYY.
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			request	body		subs.CreateSub			true	"Данные для создания подписки"
//	@Success		201		{object}	map[string]interface{}	"Подписка успешно создана"
//	@Failure		400		{object}	map[string]interface{}	"Неверный запрос"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/api/subscriptions [post]
//
//	@Example		request - подписка без даты окончания
//
//	{
//	  "service_name": "Netflix",
//	  "price": 799,
//	  "user_id": "123e4567-e89b-12d3-a456-426614174000",
//	  "start_date": "01-2024"
//	}
//
//	@Example		request - подписка с датой окончания
//
//	{
//	  "service_name": "Spotify",
//	  "price": 199,
//	  "user_id": "123e4567-e89b-12d3-a456-426614174000",
//	  "start_date": "01-2024",
//	  "end_date": "12-2024"
//	}
//
//	@Example		request - премиум подписка
//
//	{
//	  "service_name": "YouTube Premium",
//	  "price": 399,
//	  "user_id": "123e4567-e89b-12d3-a456-426614174000",
//	  "start_date": "06-2024"
//	}
//
//	@Example		response 201 - успешное создание
//
//	{
//	  "message": "Subscription created successfully",
//	  "subscription": {
//	    "id": 1,
//	    "service_name": "Netflix",
//	    "price": 799,
//	    "user_id": "123e4567-e89b-12d3-a456-426614174000",
//	    "start_date": "2024-01-01T00:00:00Z",
//	    "end_date": null,
//	    "created_at": "2024-01-15T10:30:00Z",
//	    "updated_at": "2024-01-15T10:30:00Z"
//	  }
//	}
//
//	@Example		response 400 - end_date раньше start_date
//
//	{
//	  "error": "End date cannot be before start date"
//	}
//
//	@Example		response 400 - неверный формат даты
//
//	{
//	  "error": "Invalid request body"
//	}
//
//	@Example		response 500 - ошибка создания
//
//	{
//	  "error": "Failed to create subscription"
//	}
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req subs.CreateSub
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("error binding json", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if req.EndDate != nil && req.EndDate.Before(req.StartDate.Time) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "End date cannot be before start date",
		})
		return
	}

	sub := subs.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate.Time,
		EndDate:     nil,
		UpdatedAt:   time.Now(),
	}

	if req.EndDate != nil && !req.EndDate.Time.IsZero() {
		sub.EndDate = &req.EndDate.Time
	}

	if _, err := h.service.CreateSubscription(c.Request.Context(), sub); err != nil {
		if errors.Is(err, subs.ErrFailedToCreateSubscription) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create subscription",
			})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusCreated)
}

// GetSubscription возвращает подписку по ID
//
//	@Summary		Получение подписки
//	@Description	Возвращает подписку по указанному ID
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int						true	"ID подписки"
//	@Success		200	{object}	subs.Subscription		"Подписка найдена"
//	@Failure		400	{object}	map[string]interface{}	"Неверный ID"
//	@Failure		404	{object}	map[string]interface{}	"Подписка не найдена"
//	@Router			/api/subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.log.Error("error parsing id", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscription ID",
		})
		return
	}
	sub, err := h.service.GetSubscription(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, subs.ErrFailedToFoundSubscription) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Subscription not found",
			})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, sub)
}

// GetUserSubscriptions возвращает все подписки пользователя
//
//	@Summary		Получение всех подписок пользователя
//	@Description	Возвращает список всех подписок для указанного пользователя по его ID
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path		string					true	"UUID пользователя"	Format(uuid)	Example(123e4567-e89b-12d3-a456-426614174000)
//	@Success		200		{array}		subs.Subscription		"Список подписок пользователя"
//	@Failure		400		{object}	map[string]interface{}	"Неверный формат user_id"
//	@Failure		404		{object}	map[string]interface{}	"Подписки не найдены"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/api/subscriptions/user/{user_id} [get]
func (h *SubscriptionHandler) GetUserSubscriptions(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Error("error parsing user_id", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user_id format",
		})
		return
	}
	subsList, err := h.service.GetUserSubscriptions(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, subs.ErrFailedToFoundSubscription) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Subscription not found",
			})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, subsList)
}

// UpdateSubscription обновляет существующую подписку
//
//	@Summary		Обновление подписки
//	@Description	Обновляет поля существующей подписки по ID. Можно обновить одно или несколько полей.
//	@Description	Поддерживает как PUT (полное обновление), так и PATCH (частичное обновление).
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"ID подписки"	minimum(1)	Example(1)
//	@Param			request	body		subs.UpdateSub			true	"Данные для обновления (хотя бы одно поле)"
//	@Success		200		{object}	map[string]interface{}	"Обновленная подписка"
//	@Failure		400		{object}	map[string]interface{}	"Неверный запрос (неверный ID, даты или тело запроса)"
//	@Failure		404		{object}	map[string]interface{}	"Подписка не найдена"
//	@Failure		500		{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/api/subscriptions/{id} [put]
//	@Router			/api/subscriptions/{id} [patch]
//
//	@Example		request - обновление только цены
//
//	{
//	  "price": 899
//	}
//
//	@Example		request - обновление названия сервиса
//
//	{
//	  "service_name": "Netflix Premium"
//	}
//
//	@Example		request - обновление даты окончания
//
//	{
//	  "end_date": "12-2025"
//	}
//
//	@Example		request - обновление нескольких полей
//
//	{
//	  "service_name": "Netflix Ultra HD",
//	  "price": 1299,
//	  "end_date": "12-2025"
//	}
//
//	@Example		request - обновление даты начала
//
//	{
//	  "start_date": "03-2024"
//	}
//
//	@Example		request - полное обновление
//
//	{
//	  "service_name": "Netflix Premium",
//	  "price": 899,
//	  "start_date": "01-2024",
//	  "end_date": "12-2025"
//	}
//
//	@Example		response 200 - успешное обновление
//
//	{
//	  "subscription": {
//	    "id": 1,
//	    "service_name": "Netflix Premium",
//	    "price": 899,
//	    "user_id": "123e4567-e89b-12d3-a456-426614174000",
//	    "start_date": "2024-01-01T00:00:00Z",
//	    "end_date": "2025-12-31T00:00:00Z",
//	    "updated_at": "2024-01-15T10:30:00Z"
//	  }
//	}
//
//	@Example		response 400 - неверный ID
//
//	{
//	  "error": "Invalid subscription ID"
//	}
//
//	@Example		response 400 - end_date раньше start_date
//
//	{
//	  "error": "End date cannot be before start date"
//	}
//
//	@Example		response 400 - отрицательная цена
//
//	{
//	  "error": "Invalid request body"
//	}
//
//	@Example		response 404 - подписка не найдена
//
//	{
//	  "error": "Subscription not found"
//	}
//
//	@Example		response 500 - ошибка обновления
//
//	{
//	  "error": "Internal server error"
//	}
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.log.Error("error parsing id", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscription ID",
		})
		return
	}

	existingSub, err := h.service.GetSubscription(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, subs.ErrFailedToFoundSubscription) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Subscription not found",
			})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var req subs.UpdateSub
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("error binding json", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if req.ServiceName != nil {
		existingSub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		existingSub.Price = *req.Price
	}
	if req.StartDate != nil {
		existingSub.StartDate = req.StartDate.Time
	}
	if req.EndDate != nil {
		startDate := existingSub.StartDate
		if req.StartDate != nil {
			startDate = req.StartDate.Time
		}
		if req.EndDate.Before(startDate) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "End date cannot be before start date",
			})
			return
		}
		existingSub.EndDate = &req.EndDate.Time
	}

	existingSub.UpdatedAt = time.Now()

	if _, err := h.service.UpdateSubscription(c.Request.Context(), existingSub); err != nil {
		h.log.Error("failed to update subscription", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscription": existingSub,
	})
}

// DeleteSubscription удаляет подписку
//
//	@Summary		Удаление подписки
//	@Description	Удаляет подписку по указанному ID
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"ID подписки"
//	@Success		200	"Подписка успешно удалена"
//	@Failure		400	{object}	map[string]interface{}	"Неверный ID"
//	@Failure		400	{object}	map[string]interface{}	"ID подписки не может быть отрицательным"
//	@Failure		404	{object}	map[string]interface{}	"Подписка не найдена"
//	@Failure		500	{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/api/subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.log.Error("error parsing id", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscription ID",
		})
		return
	}

	if id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscription ID",
		})
		return
	}

	if err := h.service.DeleteSubscription(c.Request.Context(), id); err != nil {
		if errors.Is(err, subs.ErrFailedToDeleteSubscription) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to delete subscription",
			})
			return
		}
		if errors.Is(err, subs.ErrFailedToGetRowsAffected) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get rows affected",
			})
			return
		}
		if errors.Is(err, subs.ErrFailedToFoundSubscription) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Subscription not found",
			})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

// ListSubscriptions возвращает список подписок с пагинацией
//
//	@Summary		Получение списка подписок
//	@Description	Возвращает список подписок с пагинацией и фильтрацией
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			user_id		query		string					true	"ID пользователя"
//	@Param			page		query		int						false	"Номер страницы"	default(1)
//	@Param			page_size	query		int						false	"Размер страницы"	default(20)
//	@Success		200			{object}	map[string]interface{}	"Список подписок"
//	@Success		204			{object}	map[string]interface{}	"Список подписок пуст"
//	@Failure		400			{object}	map[string]interface{}	"Неверный запрос"
//	@Failure		500			{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/api/subscriptions/list [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.log.Error("error parsing user_id", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user_id format",
		})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		h.log.Error("error parsing page", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid page format",
		})
		return
	}

	if page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Page must be greater than 0",
		})
		return
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if err != nil {
		h.log.Error("error parsing page_size", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid page_size format",
		})
		return
	}

	subsList, total, err := h.service.ListSubscriptions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		if errors.Is(err, subs.ErrFailedToCountSubscriptions) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "failed to count subscriptions",
			})
			return
		}
		if errors.Is(err, subs.ErrFailedToListSubscriptions) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "failed to list subscriptions",
			})
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(subsList) == 0 {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subsList,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + pageSize - 1) / pageSize,
			"has_next":    page < (total+pageSize-1)/pageSize,
			"has_prev":    page > 1,
		},
	})
}

type TotalCostRequest struct {
	UserID      string `form:"user_id" json:"user_id"`
	ServiceName string `form:"service_name" json:"service_name"`
	StartDate   string `form:"start_date" json:"start_date" binding:"required"`
	EndDate     string `form:"end_date" json:"end_date" binding:"required"`
}

// GetTotalCost возвращает общую стоимость подписок за период
//
//	@Summary		Расчет общей стоимости
//	@Description	Возвращает общую стоимость подписок за указанный период с фильтрацией
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			start_date		query		string					true	"Дата начала (MM-YYYY)"
//	@Param			end_date		query		string					true	"Дата окончания (MM-YYYY)"
//	@Param			user_id			query		string					false	"ID пользователя"
//	@Param			service_name	query		string					false	"Название сервиса"
//	@Success		200				{object}	map[string]interface{}	"Общая стоимость"
//	@Failure		400				{object}	map[string]interface{}	"Неверный запрос"
//	@Failure		500				{object}	map[string]interface{}	"Внутренняя ошибка сервера"
//	@Router			/api/subscriptions/total [get]
func (h *SubscriptionHandler) GetTotalCost(c *gin.Context) {
	var req TotalCostRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Error("error binding query", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		h.log.Error("error parsing start_date", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start_date format, use YYYY-MM",
		})
		return
	}

	endDate, err := time.Parse("01-2006", req.EndDate)
	if err != nil {
		h.log.Error("error parsing end_date", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end_date format, use YYYY-MM",
		})
		return
	}

	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "End date cannot be before start date",
		})
		return
	}

	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var userID uuid.UUID
	if req.UserID != "" {
		userID, err = uuid.Parse(req.UserID)
		if err != nil {
			h.log.Error("invalid user_id format", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user_id format",
			})
			return
		}
	}

	serviceReq := subs.CostCalculation{
		UserID:      userID,
		ServiceName: req.ServiceName,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	cost, err := h.service.GetTotalCostByPeriod(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to calculate total cost",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_cost": cost,
	})
}
