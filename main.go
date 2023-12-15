package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	file, _ := os.ReadFile("recipes.json")
	json.Unmarshal(file, &recipes)
}

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

func NewRecipeHandler(ctx *gin.Context) {
	var recipe Recipe
	if err := ctx.ShouldBindJSON(&recipe); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	ctx.JSON(http.StatusOK, recipe)
}

func FetchAllRecipesHandlers(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, recipes)
}

func UpdateRecipeHandler(ctx *gin.Context) {
	var recipe Recipe
	id := ctx.Param("id")
	if err := ctx.ShouldBindJSON(&recipe); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	for i, v := range recipes {
		if v.ID == id {
			recipes[i] = recipe
			ctx.JSON(http.StatusOK, gin.H{
				"message": "Recipe updated sucseedfully",
				"recipe":  recipe,
			})
			return
		}
	}

	ctx.JSON(http.StatusNotFound, gin.H{
		"error": "No Recipe present with the given id",
	})
}

func DeleteRecipeHandler(ctx *gin.Context) {
	id := ctx.Param("id")

	for i, v := range recipes {
		if v.ID == id {
			recipes = append(recipes[:i], recipes[i+1:]...)
			ctx.JSON(http.StatusOK, gin.H{
				"message": "Recipe Deleted successfully",
			})
			return
		}
	}

	ctx.JSON(http.StatusNotFound, gin.H{
		"message": "Recipe not found",
	})
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", FetchAllRecipesHandlers)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.Run()
}
