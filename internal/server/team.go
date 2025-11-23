package server

import (
	"github.com/vnchk1/pr-manager/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CreateTeamRequest struct {
	TeamName string        `json:"team_name"`
	Members  []models.User `json:"members"`
}

type CreateTeamResponse struct {
	BaseResponse
	Team *models.Team `json:"team,omitempty"`
}

type GetTeamResponse struct {
	BaseResponse
	*models.Team
}

func (s *Server) createTeam(c echo.Context) error {
	var req CreateTeamRequest
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
	if req.TeamName == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Название команды обязательно",
			},
		})
	}

	// Проверка уникальности ID пользователей
	memberIDs := make(map[string]bool)
	for _, member := range req.Members {
		if member.ID == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				BaseResponse: BaseResponse{
					Success: false,
					Message: "ID участника команды не может быть пустым",
				},
			})
		}
		if memberIDs[member.ID] {
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				BaseResponse: BaseResponse{
					Success: false,
					Message: "Обнаружены дублирующиеся ID участников",
				},
			})
		}
		memberIDs[member.ID] = true
	}

	team := &models.Team{
		Name:    req.TeamName,
		Members: make([]*models.User, len(req.Members)),
	}

	for i, member := range req.Members {
		team.Members[i] = &models.User{
			ID:       member.ID,
			Username: member.Username,
			TeamName: req.TeamName,
			IsActive: member.IsActive,
		}
	}

	createdTeam, err := s.service.Team.Create(c.Request().Context(), team)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Не удалось создать команду",
			},
			Error: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, CreateTeamResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: "Команда успешно создана",
		},
		Team: createdTeam,
	})
}

func (s *Server) getTeam(c echo.Context) error {
	teamName := c.QueryParam("team_name")
	if teamName == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Параметр team_name обязателен",
			},
		})
	}

	team, err := s.service.Team.Get(c.Request().Context(), teamName)
	if err != nil {

		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Не удалось получить информацию о команде",
			},
			Error: err.Error(),
		})
	}

	if team == nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			BaseResponse: BaseResponse{
				Success: false,
				Message: "Команда не найдена",
			},
		})
	}

	return c.JSON(http.StatusOK, GetTeamResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: "Информация о команде успешно получена",
		},
		Team: team,
	})
}
