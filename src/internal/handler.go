package internal

import (
	"bookem-notification-service/client/userclient"
	"bookem-notification-service/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Route struct{ handler Handler }

func NewRoute(handler Handler) *Route { return &Route{handler} }

func (r *Route) Route(rg *gin.RouterGroup) {
	rg.POST("/new", r.handler.createNotification)
}

type Handler struct{ service Service }

func NewHandler(s Service) Handler { return Handler{s} }

func (h *Handler) createNotification(ctx *gin.Context) {
	jwt, err := util.GetJwt(ctx)
	if err != nil {
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != userclient.Guest && jwt.Role != userclient.Host {
		AbortError(ctx, ErrUnauthorized)
		return
	}

	var dto NotificationDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		AbortError(ctx, err)
		return
	}

	notification, err := h.service.Create(ctx, jwt.ID, dto)
	if err != nil {
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, notification)
}
