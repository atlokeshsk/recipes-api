package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/atlokeshsk/recipes-api/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandlers struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewRecipesHandlers(collection *mongo.Collection, ctx context.Context) *RecipesHandlers {
	return &RecipesHandlers{collection: collection, ctx: ctx}
}

func (handler *RecipesHandlers) FetchAllRecipesHandlers(c *gin.Context) {
	cur, err := handler.collection.Find(handler.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	defer cur.Close(handler.ctx)
	recipes := make([]models.Recipe, 0)
	cur.All(handler.ctx, &recipes)
	c.JSON(http.StatusOK, recipes)
}

func (handler *RecipesHandlers) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

func (handler *RecipesHandlers) UpdateRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	id := c.Param("id")
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{
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

func (handler *RecipesHandlers) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.DeleteOne(handler.ctx, bson.D{{Key: "_id", Value: objectId}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "recipe deleted success fully",
	})
}

func (handler *RecipesHandlers) SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")
	recipesResult := make([]models.Recipe, 0)

	cur, err := handler.collection.Find(handler.ctx, bson.D{
		{Key: "tags", Value: tag},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
		return
	}
	defer cur.Close(handler.ctx)
	_ = cur.All(handler.ctx, &recipesResult)
	c.JSON(http.StatusOK, recipesResult)
}

func (handler *RecipesHandlers) GetRecipeByID(c *gin.Context) {
	id := c.Param("id")
	objectID, _ := primitive.ObjectIDFromHex(id)
	var recipe models.Recipe
	err := handler.collection.FindOne(handler.ctx, bson.D{
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
