package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"html/template"
	"miaro-schedule-api/pkg"
	"net/http"
)

// Embed the templates directory
//
//go:embed templates/*
var templatesFS embed.FS

func SchedulerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		schedule := pkg.CalculateSchedule()
		scheduleBeautified := pkg.FormatScheduleBeautified(schedule)

		c.HTML(http.StatusOK, "miaroSchedule.tmpl", gin.H{
			"Schedule":               scheduleBeautified.Schedule,
			"IsWorking":              scheduleBeautified.IsWorking,
			"NextWorkingDay":         scheduleBeautified.NextWorkingDay,
			"ScheduleNextWorkingDay": scheduleBeautified.ScheduleNextWorkingDay,
		})
	}
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	// Parse the templates from the embedded filesystem
	tmpl := template.Must(template.New("").ParseFS(templatesFS, "templates/*.tmpl"))
	router.SetHTMLTemplate(tmpl)

	// Define your routes
	router.GET("/miaro", SchedulerHandler())

	return router
}

func main() {
	router := setupRouter()

	// Start the server
	err := router.Run(":8081")
	if err != nil {
		panic(err)
	}
}
