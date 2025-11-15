package handlers

import (
	"net/http"
	"pr-manage-service/internal/domain"
	"pr-manage-service/internal/interfaces/dto"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	usecase domain.UserService
}

func NewUserHandler(usecase domain.UserService) *UserHandler {
	return &UserHandler{
		usecase: usecase,
	}
}

func (h *UserHandler) SetIsActiveHandler(c *gin.Context) {
	var user dto.UserRequest
	if err := c.BindJSON(&user); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if username, err := h.usecase.SetIsActive(user.TeamName, user.UserID, user.IsActive); err != nil {
		switch err.(type) {
		}
	} else {
		c.JSON(http.StatusOK, dto.UserResponse{
			UserRequest: &user,
			UserName:    username,
		})
		return
	}
}

func (h *UserHandler) GetReviewHandler(c *gin.Context) {

}
