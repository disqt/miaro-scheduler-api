package main

import (
	"github.com/gin-gonic/gin"
	"miaro-schedule-api/pkg"
	"net/http"
)

func SchedulerHandler() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		schedule := pkg.CalculateSchedule()

		scheduleBeautified := pkg.FormatScheduleBeautified(schedule)

		c.HTML(http.StatusOK, "miaroSchedule.tmpl", gin.H{
			"Schedule":               scheduleBeautified.Schedule,
			"IsWorking":              scheduleBeautified.IsWorking,
			"NextWorkingDay":         scheduleBeautified.NextWorkingDay,
			"ScheduleNextWorkingDay": scheduleBeautified.ScheduleNextWorkingDay,
		})
	}

	return fn
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.LoadHTMLFiles(
		"templates/miaroSchedule.tmpl",
		"templates/index.html",
	)
	router.GET("/miaro", SchedulerHandler())
	return router
}

func main() {
	router := setupRouter()
	err := router.Run(":8081")

	if err != nil {
		panic(err)
	}
}
