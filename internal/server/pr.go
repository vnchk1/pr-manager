package server

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type CreatePRResponse struct {
	BaseResponse
	PR *models.PullRequest `json:"pr,omitempty"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type MergePRResponse struct {
	BaseResponse
	PR *models.PullRequest `json:"pr,omitempty"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type ReassignReviewerResponse struct {
	BaseResponse
	PR         *models.PullRequest `json:"pr,omitempty"`
	ReplacedBy string              `json:"replaced_by,omitempty"`
}

func (s *Server) createPR(c echo.Context) error {
	var req CreatePRRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Неверный формат запроса",
			},
			Error: err.Error(),
		})
	}

	// Валидация обязательных полей
	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Все поля обязательны для заполнения",
			},
		})
	}

	createReq := &models.PRCreateRequest{
		ID:       req.PullRequestID,
		Name:     req.PullRequestName,
		AuthorID: req.AuthorID,
	}

	pr, err := s.service.PR.Create(c.Request().Context(), createReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Не удалось создать pull request",
			},
			Error: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, CreatePRResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: "Pull request успешно создан",
		},
		PR: pr,
	})
}

func (s *Server) mergePR(c echo.Context) error {
	var req MergePRRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Неверный формат запроса",
			},
			Error: err.Error(),
		})
	}

	if req.PullRequestID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "ID pull request обязателен",
			},
		})
	}

	mergeReq := &models.PRMergeRequest{
		ID: req.PullRequestID,
	}

	pr, err := s.service.PR.Merge(c.Request().Context(), mergeReq)
	if err != nil {
		// Можно добавить более детальную обработку разных типов ошибок
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Не удалось выполнить merge pull request",
			},
			Error: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, MergePRResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: "Pull request успешно объединен",
		},
		PR: pr,
	})
}

func (s *Server) reassignReviewer(c echo.Context) error {
	var req ReassignReviewerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Неверный формат запроса",
			},
			Error: err.Error(),
		})
	}

	if req.PullRequestID == "" || req.OldUserID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "ID pull request и старого ревьювера обязательны",
			},
		})
	}

	reassignReq := &models.PRReassignRequest{
		ID:          req.PullRequestID,
		OldReviewer: req.OldUserID,
	}

	pr, newReviewerID, err := s.service.PR.ReassignReviewer(c.Request().Context(), reassignReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Не удалось перераспределить ревьювера",
			},
			Error: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, ReassignReviewerResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: "Ревьювер успешно перераспределен",
		},
		PR:         pr,
		ReplacedBy: newReviewerID,
	})
}
