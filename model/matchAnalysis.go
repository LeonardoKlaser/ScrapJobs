package model

type ResumeAnalysis struct {
	MatchAnalysis               MatchAnalysis `json:"matchAnalysis"`
	AtsKeywords                 AtsKeywords   `json:"atsKeywords"`
	StrengthsForThisJob         []Strength    `json:"strengthsForThisJob"`
	GapsAndImprovementAreas     []Gap         `json:"gapsAndImprovementAreas"`
	ActionableResumeSuggestions []Suggestion  `json:"actionableResumeSuggestions"`
	FinalConsiderations         string        `json:"finalConsiderations"`
}

type MatchAnalysis struct {
	OverallScoreNumeric     int    `json:"overallScoreNumeric"`
	OverallScoreQualitative string `json:"overallScoreQualitative"`
	Summary                 string `json:"summary"`
}

type AtsKeywords struct {
	Matched []string `json:"matched"`
	Missing []string `json:"missing"`
}

type Strength struct {
	Point          string `json:"point"`
	RelevanceToJob string `json:"relevanceToJob"`
}

type Gap struct {
	AreaDescription      string `json:"areaDescription"`
	JobRequirementImpacted string `json:"jobRequirementImpacted"`
}

type Suggestion struct {
	Suggestion               string `json:"suggestion"`
	CurriculumSectionToApply string `json:"curriculumSectionToApply"`
	ExampleWording           string `json:"exampleWording"`
	ReasoningForThisJob      string `json:"reasoningForThisJob"`
}