package routes

import (
	"net/http"
	"project-manager/src/controllers"
	"project-manager/src/middleware"
)

// register all application endpoints
func SetupRoutes(authCtrl *controllers.AuthController, projectCtrl *controllers.ProjectController, taskCtrl *controllers.TaskController) http.Handler {
	router := http.NewServeMux()

	// base sanity check
	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "pong", "status": "running"}`))
	})

	// public Auth Endpoints
	router.HandleFunc("/api/auth/register", authCtrl.Register)
	router.HandleFunc("/api/auth/login", authCtrl.Login)

	// protected User Endpoints (User list/search)
	router.Handle("/api/users", middleware.AuthMiddleware(http.HandlerFunc(authCtrl.ListUsers)))

	// protected Project Endpoints
	router.Handle("/api/projects", middleware.AuthMiddleware(http.HandlerFunc(projectCtrl.Dispatch)))
	router.Handle("/api/projects/", middleware.AuthMiddleware(http.HandlerFunc(projectCtrl.Dispatch)))

	// protected Task Endpoints
	router.Handle("/api/tasks", middleware.AuthMiddleware(http.HandlerFunc(taskCtrl.Dispatch)))
	router.Handle("/api/tasks/", middleware.AuthMiddleware(http.HandlerFunc(taskCtrl.Dispatch)))

	return middleware.CORSMiddleware(router)
}
