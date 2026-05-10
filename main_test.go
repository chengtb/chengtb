package main

import (
"strings"
"testing"
)

func TestValidateInputSuccess(t *testing.T) {
input := RecipeCardInput{
Title:            "Garlic-Fragrant Romaine Lettuce",
RecipeNo:         "00015491",
DevelopmentDate:  "2024-01-01",
Machine:          "Cheetah (H22)",
Description:      "Healthy green dish",
PreparationSteps: []string{"Step 1", "Step 2"},
Compartments: map[string][]string{
"a": {"Garlic: 15g"},
"B": {"Lettuce: 250g", "Lard: 10g"},
"C": {"Lettuce: 100g"},
"D": {"Lettuce: 100g"},
},
IngredientsImageURL: "https://example.com/ingredients.jpg",
DishImageURL:        "https://example.com/dish.png",
}

if err := validateInput(&input); err != nil {
t.Fatalf("validateInput() error = %v", err)
}

if _, ok := input.Compartments["A"]; !ok {
t.Fatalf("expected compartments to normalize labels to uppercase")
}
}

func TestValidateInputMissingCompartment(t *testing.T) {
input := RecipeCardInput{
Title:              "Dish",
RecipeNo:           "001",
DevelopmentDate:    "2024-01-01",
Machine:            "H22",
Description:        "desc",
PreparationSteps:   []string{"step"},
Compartments:       map[string][]string{"A": {"x"}, "B": {"y"}, "C": {"z"}},
IngredientsImageURL: "https://example.com/ingredients.jpg",
DishImageURL:       "https://example.com/dish.jpg",
}

err := validateInput(&input)
if err == nil {
t.Fatal("validateInput() expected error, got nil")
}
if !strings.Contains(err.Error(), "compartments.D") {
t.Fatalf("expected missing compartments.D in error, got %v", err)
}
}

func TestValidateInputInvalidURL(t *testing.T) {
input := RecipeCardInput{
Title:            "Dish",
RecipeNo:         "001",
DevelopmentDate:  "2024-01-01",
Machine:          "H22",
Description:      "desc",
PreparationSteps: []string{"step"},
Compartments: map[string][]string{
"A": {"1"},
"B": {"1"},
"C": {"1"},
"D": {"1"},
},
IngredientsImageURL: "ftp://example.com/ingredients.jpg",
DishImageURL:        "https://example.com/dish.jpg",
}

err := validateInput(&input)
if err == nil {
t.Fatal("validateInput() expected error, got nil")
}
if !strings.Contains(err.Error(), "ingredients_image_url") {
t.Fatalf("expected ingredients_image_url error, got %v", err)
}
}
