package main

import (
	"log"
	"net/http"

	"project-manager/src/controllers"
	"project-manager/src/models"
	"project-manager/src/repositories"
	"project-manager/src/routes"
)

func main() {
	log.Println("Starting Project Management API server...")

	ConnectDatabase()

	log.Println("Running schema migrations...")
	err := DB.AutoMigrate(&models.User{}, &models.Project{}, &models.Task{})
	if err != nil {
		log.Fatal("Database migration failed: ", err)
	}
	log.Println("Database migration completed successfully.")

	userRepo := repositories.NewUserRepository(DB)
	projectRepo := repositories.NewProjectRepository(DB)

	authController := controllers.NewAuthController(userRepo)
	projectController := controllers.NewProjectController(projectRepo)

	appRouter := routes.SetupRoutes(authController, projectController)

	serverAddr := ":8080"
	log.Printf("API engine ready and listening on port %s", serverAddr)

	err = http.ListenAndServe(serverAddr, appRouter)
	if err != nil {
		log.Fatal("Failed to start HTTP server: ", err)
	}
}
