package routes

import (
	"net/http"
	"project-manager/src/controllers"
)

// register all application endpoints onto a unified multiplexer
func SetupRoutes(authCtrl *controllers.AuthController) *http.ServeMux {
	router := http.NewServeMux()

	// base sanity check
	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "pong", "status": "running"}`))
	})

	// authenticated resources mapping
	router.HandleFunc("/api/auth/register", authCtrl.Register)
	router.HandleFunc("/api/auth/login", authCtrl.Login)

	return router
}
