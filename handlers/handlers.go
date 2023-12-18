package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/atlokeshsk/recipes-api/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandlers struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewRecipesHandlers(collection *mongo.Collection, ctx context.Context, redisClient *redis.Client) *RecipesHandlers {
	return &RecipesHandlers{collection, ctx, redisClient}
}

func (handler *RecipesHandlers) FetchAllRecipesHandlers(c *gin.Context) {
	recipes := make([]models.Recipe, 0)
	val, err := handler.redisClient.Get(handler.ctx, "recipes").Result()

	if err == redis.Nil {
		log.Println("Request to mongodb")

		//Request to mongodb
		cur, err := handler.collection.Find(handler.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err,
			})
			return
		}
		defer cur.Close(handler.ctx)

		cur.All(handler.ctx, &recipes)
		data, _ := json.Marshal(recipes)
		handler.redisClient.Set(handler.ctx, "recipes", string(data), 0)
		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err,
		})
	} else {
		log.Println("Data read from redis")
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}

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
	log.Println("remove data from redis")
	handler.redisClient.Del(handler.ctx, "recipes")
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

	log.Println("remove data from redis")
	handler.redisClient.Del(handler.ctx, "recipes")
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
	log.Println("remove data from redis")
	handler.redisClient.Del(handler.ctx, "recipes")
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

	val, err := handler.redisClient.Get(handler.ctx, "recipes").Result()

	if err == redis.Nil {
		// no cached data is redis. so read from the mongododb
		var recipe models.Recipe
		err = handler.collection.FindOne(handler.ctx, bson.D{
			{Key: "_id", Value: objectID},
		}).Decode(&recipe)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err,
			})
			return
		}
		c.JSON(http.StatusOK, recipe)
	} else if err != nil {
		// error in redis read
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error in redis",
		})
	} else {
		// get data from redis
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)

		for _, recipe := range recipes {
			if recipe.ID == objectID {
				c.JSON(http.StatusOK, recipe)
				return
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error in redis",
		})
	}

}
