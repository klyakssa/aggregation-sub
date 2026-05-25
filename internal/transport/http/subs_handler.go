package http

import (
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
	service subs.SubscriptionService
}

func NewSubscriptionHandler(log *logger.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		log: log,
	}
}

func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req subs.CreateSub
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("error binding json", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if req.EndDate != nil && req.EndDate.Before(req.StartDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "End date cannot be before start date",
		})
		return
	}

	sub := subs.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		UpdatedAt:   time.Now(),
	}

	if _, err := h.service.CreateSubscription(sub); err != nil {
		h.log.Error("error creating subscription", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscription ID",
		})
		return
	}
	sub, err := h.service.GetSubscription(id)
	if err != nil {
		h.log.Error("error getting subscription", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, sub)
}

func (h *SubscriptionHandler) GetUserSubscriptions(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user_id format",
		})
		return
	}
	subs, err := h.service.GetUserSubscriptions(userID)
	if err != nil {
		h.log.Error("error getting user subscriptions", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, subs)
}

func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscription ID",
		})
		return
	}

	sub, err := h.service.GetSubscription(id)
	if err != nil {
		h.log.Error("error getting subscription", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var req subs.UpdateSub
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("error binding json", zap.Error(err))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	updates := make(map[string]interface{})

	if req.ServiceName != nil {
		updates["service_name"] = *req.ServiceName
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.StartDate != nil {
		updates["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		startDate := time.Now()
		if req.StartDate != nil {
			startDate = *req.StartDate
		}
		if req.EndDate != nil && req.EndDate.Before(startDate) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "End date cannot be before start date",
			})
			return
		}
		updates["end_date"] = req.EndDate
	}

	updates["updated_at"] = time.Now()

	if _, err := h.service.UpdateSubscription(sub); err != nil {
		h.log.Error("error updating subscription", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user_id format",
		})
		return
	}

	serviceName := c.Query("service_name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "service_name is required",
		})
		return
	}

	if err := h.service.DeleteSubscription(userID, serviceName); err != nil {
		h.log.Error("error deleting subscription", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user_id format",
		})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid page_size format",
		})
		return
	}

	subs, total, err := h.service.ListSubscriptions(userID, page, pageSize)
	if err != nil {
		h.log.Error("error listing subscriptions", zap.Error(err))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": subs,
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
