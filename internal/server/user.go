package server

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

type SetUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type SetUserActiveResponse struct {
	BaseResponse
	User *models.User `json:"user,omitempty"`
}

type GetUserReviewResponse struct {
	BaseResponse
	UserID       string                     `json:"user_id"`
	PullRequests []*models.PullRequestShort `json:"pull_requests,omitempty"`
}

func (s *Server) setUserActive(c echo.Context) error {
	var req SetUserActiveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Неверный формат запроса",
			},
			Error: err.Error(),
		})
	}

	if req.UserID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "ID пользователя обязательно",
			},
		})
	}

	user, err := s.service.User.SetActive(c.Request().Context(), req.UserID, req.IsActive)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Не удалось обновить статус пользователя",
			},
			Error: err.Error(),
		})
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Пользователь не найден",
			},
		})
	}

	statusMessage := "деактивирован"
	if req.IsActive {
		statusMessage = "активирован"
	}

	return c.JSON(http.StatusOK, SetUserActiveResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: "Пользователь успешно " + statusMessage,
		},
		User: user,
	})
}

func (s *Server) getUserReviewPRs(c echo.Context) error {
	userID := c.QueryParam("user_id")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Параметр user_id обязателен",
			},
		})
	}

	prs, err := s.service.PR.GetByReviewer(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Не удалось получить список pull requests для ревью",
			},
			Error: err.Error(),
		})
	}

	if prs == nil {
		prs = []*models.PullRequestShort{}
	}

	message := "Найдено " + string(rune(len(prs))) + " pull requests для ревью"
	if len(prs) == 0 {
		message = "Нет pull requests для ревью"
	}

	return c.JSON(http.StatusOK, GetUserReviewResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: message,
		},
		UserID:       userID,
		PullRequests: prs,
	})
}
