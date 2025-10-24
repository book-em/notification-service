package internal

import (
	"bookem-notification-service/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Route struct{ handler Handler }

func NewRoute(handler Handler) *Route { return &Route{handler} }

func (r *Route) Route(rg *gin.RouterGroup) {
	rg.POST("/notification", r.handler.createNotification)
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

	var dto NewNotificationDTO
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

	ctx.JSON(http.StatusCreated, notification)
}
