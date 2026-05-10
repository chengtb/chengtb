package main

import "testing"

func TestValidateRequiresAllBins(t *testing.T) {
	in := RecipeCardInput{
		RecipeName:       "test",
		RecipeNumber:     "001",
		DevelopmentTime:  "2026-05-10",
		MachineType:      "X1",
		Description:      "desc",
		PreparationSteps: []string{"step1"},
		Bins: map[string]BinIngredient{
			"A": {Ingredient: "a"},
			"B": {Ingredient: "b"},
			"C": {Ingredient: "c"},
		},
		RawMaterialImage: "/tmp/raw.png",
		FinishedImage:    "/tmp/done.png",
	}
	if err := in.Validate(); err == nil {
		t.Fatalf("expected validation error when bin D is missing")
	}
}

func TestFormatStepsSkipsBlankLines(t *testing.T) {
	got := formatSteps([]string{" first ", "", "second"})
	want := "1. first\n2. second"
	if got != want {
		t.Fatalf("unexpected formatted steps:\nwant: %q\ngot:  %q", want, got)
	}
}
