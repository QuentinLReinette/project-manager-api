package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"net/http"
)

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:191;unique;not null"`
	Password  string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func main() {
	log.Println("Démarrage de l'API de gestion de projet...")

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbName == "" {
		log.Fatal("Erreur : Les variables d'environnement de la base de données ne sont pas complètes.")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName,
	)

	var db *gorm.DB
	var err error

	log.Println("Connexion à la base de données MySQL...")
	for i := 1; i <= 5; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("[Tentative %d/5] Échec de la connexion à MySQL, nouvelle tentative dans 3 secondes...", i)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatal("Erreur critique : Impossible de se connecter à la base de données après 5 tentatives :", err)
	}

	log.Println("Connexion réussie à MySQL via GORM !")

	log.Println("Lancement de la migration des tables...")
	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal("Erreur lors de la migration de la base de données :", err)
	}
	log.Println("Migration terminée avec succès. La base de données est à jour.")

	router := http.NewServeMux()

	router.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "pong", "status": "running"}`))
	})

	serverAddr := ":8080"
	log.Printf("API prête et à l'écoute sur le port %s 🚀", serverAddr)

	err = http.ListenAndServe(serverAddr, router)
	if err != nil {
		log.Fatal("Échec du démarrage du serveur HTTP :", err)
	}
}
