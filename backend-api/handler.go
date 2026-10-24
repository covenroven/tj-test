package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type handler struct {
	service *vehicleService
}

func newHandler(svc *vehicleService) *handler {
	return &handler{
		service: svc,
	}
}

func (h *handler) getLatestVehicleLocation(c *gin.Context) {
	vehicleId := c.Param("vehicleId")
	if vehicleId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter"})
		return
	}

	loc, err := h.service.GetLatestLocation(c.Request.Context(), vehicleId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": loc,
	})
}

func (h *handler) getVehicleLocationHistory(c *gin.Context) {
	vehicleId := c.Param("vehicleId")
	if vehicleId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter"})
		return
	}

	start := c.Query("start")
	timeStart, err := strconv.ParseInt(start, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameter"})
		return
	}

	end := c.Query("end")
	timeEnd, err := strconv.ParseInt(end, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameter"})
		return
	}

	res, err := h.service.GetLocationsByRange(c.Request.Context(), vehicleId, timeStart, timeEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": res,
	})
}

func (h *handler) registerRoute(s *gin.Engine) {
	s.GET("/vehicles/:vehicleId/location", h.getLatestVehicleLocation)
	s.GET("/vehicles/:vehicleId/history", h.getVehicleLocationHistory)
}
