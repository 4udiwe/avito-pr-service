package post_user_is_active

import (
	"errors"
	"net/http"

	api "github.com/4udiwe/avito-pr-service/internal/api/http"
	"github.com/4udiwe/avito-pr-service/internal/api/http/decorator"
	"github.com/4udiwe/avito-pr-service/internal/dto"
	service "github.com/4udiwe/avito-pr-service/internal/service/user"
	"github.com/labstack/echo/v4"
)

type handler struct {
	s UserService
}

func New(userService UserService) api.Handler {
	return decorator.NewBindAndValidateDerocator(&handler{s: userService})
}

type Request dto.PostUsersSetIsActiveJSONBody

func (h *handler) Handle(ctx echo.Context, in Request) error {
	user, err := h.s.SetUserStatus(ctx.Request().Context(), in.UserId, in.IsActive)

	if err != nil {
		var errResponse dto.ErrorResponse

		if errors.Is(err, service.ErrUserNotFound) {
			errResponse.Error.Code = dto.NOTFOUND
			errResponse.Error.Message = "resource not found"
			return echo.NewHTTPError(http.StatusNotFound, errResponse)
		}

		errResponse.Error.Message = err.Error()
		return echo.NewHTTPError(http.StatusInternalServerError, errResponse)
	}

	response := struct {
		User dto.User `json:"user"`
	}{}
	response.User.FillFromEntity(user)

	return ctx.JSON(http.StatusOK, response)
}
