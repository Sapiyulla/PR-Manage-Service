package handlers

import (
	"context"
	"net/http"
	"pr-manage-service/internal/domain"
	"pr-manage-service/internal/interfaces/dto"
	"pr-manage-service/pkg/codes"
	"pr-manage-service/pkg/errs"

	"github.com/gin-gonic/gin"
)

type PrHandler struct {
	usecase domain.PRService
}

func NewPRHandler(ctx context.Context, repo domain.PRService) *PrHandler {
	prHandler := &PrHandler{
		usecase: repo,
	}
	return prHandler
}

func (h *PrHandler) CreateHandler(c *gin.Context) {
	var req dto.PRCreateRequest
	if err := c.BindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if resp, err := h.usecase.Create(&req); err != nil {
		switch err.(type) {
		case *errs.NotFoundError:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Err: dto.ErrorResponseBody{
					Code: codes.NOT_FOUND,
					Msg:  err.Error(),
				},
			})
			return
		case *errs.InternalError:
			c.Status(http.StatusInternalServerError)
			return
		case *errs.AlreadyExistsError:
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Err: dto.ErrorResponseBody{
					Code: codes.PR_EXISTS,
					Msg:  err.Error(),
				},
			})
		}
	} else {
		c.JSON(http.StatusCreated, resp)
		return
	}
}

func (h *PrHandler) MergeHandler(c *gin.Context) {
	var mergeReq dto.PRCreateRequest
	if err := c.BindJSON(&mergeReq); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if resp, err := h.usecase.Merge(&mergeReq); err != nil {
		switch err.(type) {
		case *errs.NotFoundError:
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Err: dto.ErrorResponseBody{
					Code: codes.NOT_FOUND,
					Msg:  err.Error(),
				},
			})
			return
		case *errs.InternalError:
			c.Status(http.StatusInternalServerError)
		}
	} else {
		c.JSON(http.StatusOK, gin.H{`pr`: resp})
		return
	}
}

func (h *PrHandler) ReassignHandler(c *gin.Context) {
	var req struct {
		PrID     string `json:"pull_request_id"`
		OldRevID string `json:"old_reviewer_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if _, err := h.usecase.Reassign(req.PrID, req.OldRevID); err != nil {
		switch err.(type) {

		}
	} else {

	}
}
