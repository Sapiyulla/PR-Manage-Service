package handlers

import (
	"net/http"
	"pr-manage-service/internal/domain"
	"pr-manage-service/internal/interfaces/dto"
	"pr-manage-service/pkg/codes"
	"pr-manage-service/pkg/errs"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	usecase    domain.UserService
	adminToken string
}

func NewUserHandler(usecase domain.UserService, adminToken string) *UserHandler {
	return &UserHandler{
		usecase:    usecase,
		adminToken: adminToken,
	}
}

func (h *UserHandler) SetIsActiveHandler(c *gin.Context) {
	if c.GetHeader("Admin-Token") != h.adminToken {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Err: dto.ErrorResponseBody{
				Code: codes.NOT_FOUND,
				Msg:  "resource not found",
			},
		})
		return
	}
	var user dto.UserRequest
	if err := c.BindJSON(&user); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if username, err := h.usecase.SetIsActive(user.TeamName, user.UserID, user.IsActive); err != nil {
		switch err.(type) {
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
		c.JSON(http.StatusOK, dto.UserFullResponse{
			User: dto.UserResponse{
				UserRequest: &user,
				UserName:    username,
			},
		})
		return
	}
}

func (h *UserHandler) GetReviewHandler(c *gin.Context) {
	var teamName, userID string
	teamName, has := c.GetQuery("team_name")
	if !has {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Err: dto.ErrorResponseBody{
				Code: codes.INVALID_INPUT,
				Msg:  "indefined 'team_name' query var",
			},
		})
		return
	}
	userID, has = c.GetQuery("user_id")
	if !has {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Err: dto.ErrorResponseBody{
				Code: codes.INVALID_INPUT,
				Msg:  "indefined 'user_id' query var",
			},
		})
		return
	}
	if resp, err := h.usecase.GetReview(teamName, userID); err != nil {
		switch err.(type) {
		case *errs.InternalError:
			c.Status(http.StatusInternalServerError)
			return
		case *errs.NotFoundError:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Err: dto.ErrorResponseBody{
					Code: codes.NOT_FOUND,
				},
			})
			return
		}
	} else {
		c.JSON(http.StatusOK, resp)
	}
}
