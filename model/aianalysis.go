package model

type RequirementAnalysis struct {
	RequirementJob string `json:"requirement_job"`
	CandidateCompare string `json:"candidate_compare"`
	AnalysisResult string `json:"analysis_result"`
}

type AIAnalysisResult struct{
	MatchScore int `json:"match_score"`
	Summary string `json:"summary"`
	Strenghts []string `json:"strengths"`
	Gaps []string `json:"gaps"`
	Recommendations []RequirementAnalysis `json:"recommendations"`
}

