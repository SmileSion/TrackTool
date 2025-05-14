package controller

import (
	"net/http"
	"github.com/qwy-tacking/middleware"
	"github.com/qwy-tacking/model"
	"github.com/qwy-tacking/storage"

	"github.com/gin-gonic/gin"
)

func TrackHandler(c *gin.Context) {
	var event model.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := storage.SaveEventToRedis(event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "redis error"})
		return
	}

	middleware.Logger.Printf("收到事件: %v\n", event)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
