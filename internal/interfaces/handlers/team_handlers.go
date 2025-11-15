package handlers

import (
	"net/http"
	"pr-manage-service/internal/domain"
	"pr-manage-service/internal/interfaces/dto"
	"pr-manage-service/pkg/codes"
	"pr-manage-service/pkg/errs"

	"github.com/gin-gonic/gin"
)

type TeamHandler struct {
	usecase domain.TeamService
}

func NewTeamHandler(usecase domain.TeamService) *TeamHandler {
	return &TeamHandler{
		usecase: usecase,
	}
}

func (h *TeamHandler) AddTeamHandler(c *gin.Context) {
	var team dto.TeamRequest
	if err := c.BindJSON(&team); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if err := h.usecase.AddTeam(&team); err != nil {
		switch err.(type) {
		case *errs.InvalidError:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Err: dto.ErrorResponseBody{
					Code: codes.INVALID_INPUT,
					Msg:  err.Error(),
				},
			})
		// 500
		case *errs.InternalError:
			c.Status(http.StatusInternalServerError)
			return
		// 400
		case *errs.AlreadyExistsError:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Err: dto.ErrorResponseBody{
					Code: codes.TEAM_EXISTS,
					Msg:  err.Error(),
				},
			})
		}
	} else {
		c.JSON(http.StatusCreated, dto.TeamResponse{
			Team: team,
		})
	}
}

func (h *TeamHandler) GetTeamHandler(c *gin.Context) {
	queryTeamName, has := c.GetQuery("team_name")
	if !has {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Err: dto.ErrorResponseBody{
				Code: codes.NOT_FOUND,
				Msg:  "team not found",
			},
		})
		return
	}
	if team, err := h.usecase.GetTeamByName(queryTeamName); err != nil {
		switch err.(type) {
		case *errs.InvalidError:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Err: dto.ErrorResponseBody{
					Code: codes.INVALID_INPUT,
					Msg:  err.Error(),
				},
			})
		case *errs.InternalError:
			c.Status(http.StatusInternalServerError)
			return
		case *errs.NotFoundError:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Err: dto.ErrorResponseBody{
					Code: codes.NOT_FOUND,
					Msg:  err.Error(),
				},
			})
		}
	} else {
		c.JSON(http.StatusOK, team.Team)
		return
	}
}
