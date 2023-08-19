package enum

type ActionTypeEnum string

var ActionType = struct {
	POSITIVE ActionTypeEnum
	NEUTRAL  ActionTypeEnum
	NEGATIVE ActionTypeEnum
}{
	POSITIVE: "POSITIVE",
	NEUTRAL:  "NEUTRAL",
	NEGATIVE: "NEGATIVE",
}

type ActionCategoryEnum string

var ActionCategory = struct {
	VOTE_BLOCKING     ActionCategoryEnum
	VOTE_AVOIDING     ActionCategoryEnum
	VOTE_REDIRECTION  ActionCategoryEnum
	VOTE_IMMUNITY     ActionCategoryEnum
	VOTE_CHANGE       ActionCategoryEnum
	VISIT_BLOCKING    ActionCategoryEnum
	VISIT_REDIRECTION ActionCategoryEnum
	REACTIVE          ActionCategoryEnum
	INVESTIGATION     ActionCategoryEnum
	KILLING           ActionCategoryEnum
	PROTECTION        ActionCategoryEnum
	SUPPORT           ActionCategoryEnum
	HEALING           ActionCategoryEnum
	DEBUFF            ActionCategoryEnum
	THEFT             ActionCategoryEnum
	DESTRUCTION       ActionCategoryEnum
	ALTERATION        ActionCategoryEnum
	VISITING          ActionCategoryEnum
	GLOBAL_COOLDOWN   ActionCategoryEnum
}{
	VOTE_BLOCKING:     "VOTE_BLOCKING",
	VOTE_AVOIDING:     "VOTE_AVOIDING",
	VOTE_REDIRECTION:  "VOTE_REDIRECTION",
	VOTE_IMMUNITY:     "VOTE_IMMUNITY",
	VOTE_CHANGE:       "VOTE_CHANGE",
	VISIT_BLOCKING:    "VISIT_BLOCKING",
	VISIT_REDIRECTION: "VISIT_REDIRECTION",
	REACTIVE:          "REACTIVE",
	INVESTIGATION:     "INVESTIGATION",
	KILLING:           "KILLING",
	PROTECTION:        "PROTECTION",
	SUPPORT:           "SUPPORT",
	HEALING:           "HEALING",
	DEBUFF:            "DEBUFF",
	THEFT:             "THEFT",
	DESTRUCTION:       "DESTRUCTION",
	ALTERATION:        "ALTERATION",
	VISITING:          "VISITING",
	GLOBAL_COOLDOWN:   "GLOBAL_COOLDOWN",
}
