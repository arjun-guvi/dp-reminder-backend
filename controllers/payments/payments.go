package payments

import (
	"errors"
	"net/http"

	"github.com/ares/dp-vc-webApp/configs/response"
	"github.com/ares/dp-vc-webApp/models/payment"
	paymentservice "github.com/ares/dp-vc-webApp/services/payments"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type Controller struct{ service *paymentservice.Service }

func New(service *paymentservice.Service) *Controller { return &Controller{service: service} }

func (ctrl *Controller) Create(c *gin.Context) {
	var request payment.CreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	item, err := ctrl.service.Create(c.Request.Context(), request)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	response.Created(c, item)
}

func (ctrl *Controller) List(c *gin.Context) {
	items, err := ctrl.service.List(c.Request.Context(), c.Query("status"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, items)
}

func (ctrl *Controller) Get(c *gin.Context) {
	item, err := ctrl.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			response.NotFound(c, "Payment not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, item)
}

func (ctrl *Controller) UpdateStatus(c *gin.Context) {
	var request payment.UpdateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if err := ctrl.service.UpdateStatus(c.Request.Context(), c.Param("id"), request.Status); err != nil {
		if err.Error() == "payment not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (ctrl *Controller) Delete(c *gin.Context) {
	if err := ctrl.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		if err.Error() == "payment not found" {
			response.NotFound(c, err.Error())
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}
