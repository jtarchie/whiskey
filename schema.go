package main

type Bottle struct {
	Name                  string   `json:"name" description:"Official name of the liquor as stated on the bottle" required:"true"`
	Brand                 string   `json:"brand" description:"Brand or manufacturer of the liquor" required:"true"`
	Type                  string   `json:"type" description:"Type of liquor (e.g., Whiskey, Rum, Tequila, etc.)" required:"true"`
	Subtype               string   `json:"subtype,omitempty" description:"Subcategory (e.g., Single Malt, Reposado, Bourbon, etc.)"`
	AlcoholContent        string   `json:"alcohol_content" description:"Alcohol percentage (ABV) as stated on the label" required:"true"`
	Proof                 string   `json:"proof,omitempty" description:"Proof value of the liquor (if stated separately from ABV)"`
	Volume                string   `json:"volume" description:"Volume of the bottle (e.g., 750ml, 1L, etc.)" required:"true"`
	Origin                string   `json:"origin,omitempty" description:"Country or region of origin as per the label"`
	Distillery            string   `json:"distillery,omitempty" description:"Name of the distillery or production facility, if available"`
	BottleNumber          string   `json:"bottle_number,omitempty" description:"Unique bottle number if it's a limited edition or numbered release"`
	BatchNumber           string   `json:"batch_number,omitempty" description:"Batch number if indicated on the label"`
	Aging                 string   `json:"aging,omitempty" description:"Aging information (e.g., 12 years, 5 months, etc.)"`
	MashBill              string   `json:"mash_bill,omitempty" description:"Grain composition used in the production (e.g., 70% Corn, 15% Rye, 15% Barley)"`
	Ingredients           []string `json:"ingredients,omitempty" description:"List of ingredients if mentioned (e.g., Malted Barley, Corn, etc.)"`
	BottleShapeOrFeatures string   `json:"bottle_shape_or_features,omitempty" description:"Distinctive features of the bottle shape, material, or design elements"`
	BottleColorOrGlassColor string `json:"bottle_color_or_glass_color,omitempty" description:"Color of the bottle or the glass (e.g., Amber, Clear, Green)"`
	LabelColorScheme      string   `json:"label_color_scheme,omitempty" description:"Primary color scheme of the label (e.g., Black and Gold, Blue and Silver)"`
	LabelLanguages        []string `json:"label_languages,omitempty" description:"Languages detected on the label"`
	VisibleAwardsOrMedals []string `json:"visible_awards_or_medals,omitempty" description:"Awards or medals displayed on the bottle (e.g., Double Gold at SF World Spirits Competition)"`
	GovernmentWarnings    string   `json:"government_warnings,omitempty" description:"Legal warnings such as 'Government Warning: According to the Surgeon General...'"`
	SignatureOrSignatory  string   `json:"signature_or_signatory,omitempty" description:"Signature or signatory name if present on the label (e.g., Master Distiller John Doe)"`
	ReleaseYearOrVintage  string   `json:"release_year_or_vintage,omitempty" description:"Year of release or vintage (if specified on the label)"`
	ProductionMethod      string   `json:"production_method,omitempty" description:"Production details such as 'Non-chill filtered' or 'Pot distilled'"`
	BottlingDate          string   `json:"bottling_date,omitempty" description:"Bottling date if indicated on the label"`
	ContactInformation    string   `json:"contact_information,omitempty" description:"Manufacturer’s website, email, or customer support contact"`
	TaglineOrSlogan       string   `json:"tagline_or_slogan,omitempty" description:"Marketing tagline or slogan printed on the label"`
	BarcodeOrSerial       string   `json:"barcode_or_serial,omitempty" description:"Barcode or unique serial number if visible on the bottle"`
	CertificationsOrMarks []string `json:"certifications_or_legal_marks,omitempty" description:"Certification marks such as 'Bottled in Bond,' 'Organic,' 'Kosher,' etc."`
	BottleStory           string   `json:"bottle_story,omitempty" description:"A short description or history of the bottle if mentioned on the label"`
	AdditionalNotes       string   `json:"additional_notes,omitempty" description:"Any other relevant details that do not fit into the above categories"`
}


type BottlesSchema struct {
	Bottles []Bottle `json:"bottles" description:"List of bottles" required:"true"`
}