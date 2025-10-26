package internal

import (
	"bookem-notification-service/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Route struct{ handler Handler }

func NewRoute(handler Handler) *Route { return &Route{handler} }

func (r *Route) Route(rg *gin.RouterGroup) {
	rg.POST("/notification", r.handler.createNotification)
	rg.GET("/notifications", r.handler.getUserNotifications)
	rg.PUT("/notifications/:id/read", r.handler.markNotificationAsRead)
	rg.GET("/notifications/unread-count", r.handler.getUnreadNotificationCount)
	rg.GET("/notification/preferences", r.handler.getUserNotificationPreferences)
	rg.PUT("/notification/preferences", r.handler.updateNotificationPreferences)

}

type Handler struct{ service Service }

func NewHandler(s Service) Handler { return Handler{s} }

func (h *Handler) createNotification(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "create-notification-api")
	defer util.TEL.Pop()

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Guest && jwt.Role != util.Host {
		util.TEL.Error("user is not guest or host", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	var dto CreateNotificationDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		util.TEL.Error("failed binding JSON", err)
		AbortError(ctx, ErrBadRequestCustom("invalid request body"))
		return
	}

	notification, err := h.service.CreateNotification(util.TEL.Ctx(), jwt.ID, dto)
	if err != nil {
		util.TEL.Error("failed creating notification", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, NewNotificationDTO(notification))
}

func (h *Handler) getUserNotifications(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "get-notifications-api")
	defer util.TEL.Pop()

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	limitStr := ctx.DefaultQuery("limit", "10")
	offsetStr := ctx.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		util.TEL.Error("invalid limit query param", err, "limit", limitStr)
		AbortError(ctx, ErrBadRequestCustom("invalid limit"))
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		util.TEL.Error("invalid offset query param", err, "offset", offsetStr)
		AbortError(ctx, ErrBadRequestCustom("invalid offset"))
		return
	}

	notifications, err := h.service.GetUserNotifications(util.TEL.Ctx(), jwt.ID, limit, offset)
	if err != nil {
		util.TEL.Error("failed fetching notifications", err)
		AbortError(ctx, err)
		return
	}

	util.TEL.Debug("building response")

	result := make([]NotificationDTO, 0)
	for _, notif := range notifications {
		result = append(result, NewNotificationDTO(&notif))
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *Handler) markNotificationAsRead(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "mark-notification-as-read-api")
	defer util.TEL.Pop()

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	notificationID := ctx.Param("id")
	if notificationID == "" {
		AbortError(ctx, ErrBadRequestCustom("notification ID is required"))
		return
	}

	if err := h.service.MarkNotificationAsRead(util.TEL.Ctx(), jwt.ID, notificationID); err != nil {
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

func (h *Handler) getUnreadNotificationCount(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "get-unread-notifications-count")
	defer util.TEL.Pop()

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	count, err := h.service.GetUnreadNotificationCount(util.TEL.Ctx(), jwt.ID)
	if err != nil {
		util.TEL.Error("failed fetching unread notification count", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"unreadCount": count})
}

// ---------------------- Notification Preferences ----------------------

func (h *Handler) getUserNotificationPreferences(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "get-notification-preferences")
	defer util.TEL.Pop()

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	prefs, err := h.service.GetUserNotificationPreferences(util.TEL.Ctx(), jwt.ID)
	if err != nil {
		util.TEL.Error("failed fetching user notification preferences", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, NewNotificationPreferencesDTO(prefs))
}

func (h *Handler) updateNotificationPreferences(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "update-notification-preferences")
	defer util.TEL.Pop()

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	var dto NotificationPreferencesDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		util.TEL.Error("failed binding JSON", err)
		AbortError(ctx, ErrBadRequestCustom("invalid request body"))
		return
	}

	prefs, err := h.service.UpdateNotificationPreferences(util.TEL.Ctx(), jwt.ID, dto.EnabledTypes)
	if err != nil {
		util.TEL.Error("failed updating notification preferences", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, NewNotificationPreferencesDTO(prefs))
}
