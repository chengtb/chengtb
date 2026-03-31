package models

import "time"

type Chef struct {
	ChefID      int       `gorm:"primaryKey;autoIncrement;column:chef_id" json:"chef_id"`
	Name        string    `gorm:"column:name;not null;size:100" json:"name"`
	MaxLoad     int       `gorm:"column:max_load;default:10" json:"max_load"`
	CurrentLoad int       `gorm:"column:current_load;default:0" json:"current_load"`
	IsActive    bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Recipes []*Recipe `gorm:"many2many:chef_recipe;foreignKey:ChefID;joinForeignKey:ChefID;references:RecipeID;joinReferences:RecipeID" json:"recipes,omitempty"`
}

func (Chef) TableName() string { return "chef" }

type ChefRecipe struct {
	ID       int `gorm:"primaryKey;autoIncrement;column:id"`
	ChefID   int `gorm:"column:chef_id;not null;uniqueIndex:uk_chef_recipe"`
	RecipeID int `gorm:"column:recipe_id;not null;uniqueIndex:uk_chef_recipe"`
}

func (ChefRecipe) TableName() string { return "chef_recipe" }
