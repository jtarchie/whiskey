package main

import (
	"cmp"
	"slices"
)

type Bottle struct {
	Name                  string   `json:"name" description:"Official name of the liquor as stated on the bottle" required:"true"`
	Brand                 string   `json:"brand" description:"Brand or manufacturer of the liquor" required:"true"`
	Type                  string   `json:"type" description:"Type of liquor" required:"true"`
	Subtype               string   `json:"subtype,omitempty" description:"Subcategory (e.g., Reposado, Anejo, Single Malt, etc.)" required:"true"`
	AlcoholContent        string   `json:"alcohol_content" description:"Alcohol percentage (ABV) as stated on the label" required:"true"`
	Volume                string   `json:"volume" description:"Volume of the bottle (e.g., 750ml, 1L)" required:"true"`
	Origin                string   `json:"origin,omitempty" description:"Country of origin as per the label" required:"true"`
	Distillery            string   `json:"distillery,omitempty" description:"Name of the distillery or production facility, if available" required:"true"`
	BottleNumber          string   `json:"bottle_number,omitempty" description:"Unique bottle number if it's a limited edition or numbered bottle" required:"true"`
	BatchNumber           string   `json:"batch_number,omitempty" description:"Batch number if indicated on the label" required:"true"`
	Aging                 string   `json:"aging,omitempty" description:"Aging information (e.g., 12 years, 5 months, etc.)" required:"true"`
	Ingredients           []string `json:"ingredients,omitempty" description:"List of ingredients if mentioned" required:"true"`
	BottleShapeOrFeatures string   `json:"bottle_shape_or_features,omitempty" description:"Distinctive features of the bottle shape, material, or design elements" required:"true"`
	LabelLanguages        []string `json:"label_languages,omitempty" description:"Languages detected on the label" required:"true"`
	BarcodeOrSerial       string   `json:"barcode_or_serial,omitempty" description:"Barcode or serial number if visible on the bottle" required:"true"`
	CertificationsOrMarks []string `json:"certifications_or_legal_marks,omitempty" description:"List of any certification marks (e.g., DOC, Organic, Kosher, etc.)" required:"true"`
	BottleStory           string   `json:"bottle_story,omitempty" description:"A short description or history of the bottle if mentioned on the label" required:"true"`
	AdditionalNotes       string   `json:"additional_notes,omitempty" description:"Any other relevant details that do not fit into the above categories" required:"true"`
}

type BottlesSchema struct {
	Bottles []Bottle `json:"bottles" description:"List of bottles" required:"true"`
}

func mergeBottles(bottles []Bottle) []Bottle {
	merged := make(map[string]Bottle)

	for _, bottle := range bottles {
		key := bottle.Name + "|" + bottle.Brand
		if existing, found := merged[key]; found {
			// Merge fields if necessary, e.g., append ingredients
			existing.Ingredients = uniq(existing.Ingredients, bottle.Ingredients)
			existing.LabelLanguages = uniq(existing.LabelLanguages, bottle.LabelLanguages)
			existing.CertificationsOrMarks = uniq(existing.CertificationsOrMarks, bottle.CertificationsOrMarks)
			merged[key] = existing
		} else {
			merged[key] = bottle
		}
	}

	result := make([]Bottle, 0, len(merged))
	for _, bottle := range merged {
		result = append(result, bottle)
	}

	return result
}

func uniq[T cmp.Ordered](a []T, b []T) []T {
	c := append(a, b...)
	slices.Sort(c)
	return slices.Compact(c)
}