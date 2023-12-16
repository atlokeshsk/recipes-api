package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var recipes []Recipe
var ctx context.Context
var client *mongo.Client
var recipeCollection *mongo.Collection

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
	recipeCollection = client.Database(os.Getenv("MONGODB_DATABASE")).Collection("recipes")
}

type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := recipeCollection.InsertOne(ctx, recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

func FetchAllRecipesHandlers(c *gin.Context) {
	cur, err := recipeCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	recipes := make([]Recipe, 0)
	cur.All(ctx, &recipes)
	c.JSON(http.StatusOK, recipes)
}

func UpdateRecipeHandler(c *gin.Context) {
	var recipe Recipe
	id := c.Param("id")
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := recipeCollection.UpdateOne(ctx, bson.M{
		"_id": objectId,
	}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "name", Value: recipe.Name},
			{Key: "tags", Value: recipe.Tags},
			{Key: "ingredients", Value: recipe.Ingredients},
			{Key: "instructions", Value: recipe.Instructions},
		}},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe updated Successfully",
	})
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := recipeCollection.DeleteOne(ctx, bson.D{{Key: "_id", Value: objectId}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "recipe deleted success fully",
	})
}

func SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")
	recipesResult := make([]Recipe, 0)

	cur, err := recipeCollection.Find(ctx, bson.D{
		{Key: "tags", Value: tag},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	_ = cur.All(ctx, &recipesResult)
	c.JSON(http.StatusOK, recipesResult)
}

func GetRecipeByID(c *gin.Context) {
	id := c.Param("id")
	objectID, _ := primitive.ObjectIDFromHex(id)
	var recipe Recipe
	err := recipeCollection.FindOne(ctx, bson.D{
		{Key: "_id", Value: objectID},
	}).Decode(&recipe)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", FetchAllRecipesHandlers)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipeHandler)
	router.GET("/recipes/:id", GetRecipeByID)
	router.Run()
}
