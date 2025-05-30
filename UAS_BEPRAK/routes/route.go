package routes

import (
	"project-crud/controllers"
	"project-crud/middleware"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RouterApp(app *fiber.App) {

    // API (Format)
    api := app.Group("/api")

    // Route untuk login (ALL USER)
    api.Post("/login", controllers.Login)

    // Grup route untuk admin
    objectID, _ := primitive.ObjectIDFromHex("675d1cd023322aa0cdbdfdbd")
    adminGroup := api.Group("/admin", middleware.JWTAuth,middleware.CheckRole(objectID))
    
    // CRUD ROLE (5 ROUTE)
    adminGroup.Post("/create-roles", controllers.CreateRole)
    adminGroup.Get("/get-roles", controllers.GetRoles)
    adminGroup.Get("/get-roles/:id", controllers.GetRole)
    adminGroup.Put("/edit-roles/:id", controllers.EditRole)
    adminGroup.Delete("/delete-roles/:id", controllers.DeleteRole)

    // CRUD KATEGORI MODUL (5 ROUTE)
    adminGroup.Post("/create-kategorimoduls", controllers.CreateKategoriModul)
    adminGroup.Get("/get-kategorimoduls", controllers.GetAllKategoriModul)
    adminGroup.Get("/get-kategorimodul/:id", controllers.GetKategoriModulByID)
    adminGroup.Put("/edit-kategorimoduls/:id", controllers.EditKategoriModul)
    adminGroup.Delete("/delete-kategorimoduls/:id", controllers.DeleteKategoriModul)

    // CRUD MODUL (5 ROUTE)
    adminGroup.Post("/create-moduls", controllers.CreateModul)
    adminGroup.Get("/get-moduls", controllers.GetAllModul)
    adminGroup.Get("/get-modul/:id", controllers.GetModulByID)
    adminGroup.Put("/edit-moduls/:id", controllers.EditModul)
    adminGroup.Delete("/delete-moduls/:id", controllers.DeleteModul)

    // CRUD JENIS USER (6 ROUTE)
    adminGroup.Post("/create-jenis-user", controllers.CreateJenisUser)
    adminGroup.Get("/get-jenis-users", controllers.GetAllJenisUser)
    adminGroup.Get("/get-jenis-user/:id", controllers.GetJenisUserByID)
    adminGroup.Put("/edit-jenis-user/:id", controllers.EditJenisUser)
    adminGroup.Delete("/delete-templatemodul-jenisuser/:id", controllers.DeleteTemplateModul)
    adminGroup.Delete("/delete-jenis-user/:id", controllers.DeleteJenisUser)

    // CRUD USER (8 ROUTE)
    adminGroup.Post("/create-user", controllers.CreateUser)
    adminGroup.Get("/get-users", controllers.GetAllUsers)
    adminGroup.Get("/get-user/:id", controllers.GetUserByID)
    adminGroup.Put("/edit-user/:id", controllers.EditUser)
    adminGroup.Put("/update-jenisuser/:id", controllers.EditJenisUserFromUser)
    adminGroup.Post("/add-moduluser-tertentu", controllers.AddUserModule)
    adminGroup.Delete("/delete-moduluser-tertentu", controllers.RemoveUserModule)
    adminGroup.Delete("/delete-user/:id", controllers.DeleteUser)


    // Grup route untuk CIVITAS

}