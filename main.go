package main

import (
	"context"
	"log"
	"os"

	"github.com/atlokeshsk/recipes-api/handlers"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var ctx context.Context
var client *mongo.Client
var recipesHandler *handlers.RecipesHandlers
var redisClient *redis.Client

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	ctx = context.Background()
	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Mongodb Connected Sucessfully")
	recipeCollection := client.Database(os.Getenv("MONGODB_DATABASE")).Collection("recipes")

	// redis client initalization
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	status := redisClient.Ping(ctx)
	log.Println(status.Result())

	recipesHandler = handlers.NewRecipesHandlers(recipeCollection, ctx, redisClient)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", recipesHandler.NewRecipeHandler)
	router.GET("/recipes", recipesHandler.FetchAllRecipesHandlers)
	router.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	router.GET("/recipes/search", recipesHandler.SearchRecipeHandler)
	router.GET("/recipes/:id", recipesHandler.GetRecipeByID)
	router.Run()
}
