package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"html/template"
	"miaro-schedule-api/pkg"
	"net/http"
)

//go:embed templates/*
var templatesFS embed.FS

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

	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.tmpl"))

	router.SetHTMLTemplate(tmpl)
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
