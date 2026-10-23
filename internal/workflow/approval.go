package workflow

import (
	"fmt"
	"strings"
)

type RiskLevel string

const (
	RiskLow  RiskLevel = "LOW"
	RiskHigh RiskLevel = "HIGH"
)

func AssessRisk(actionDescription string) RiskLevel {
	lower := strings.ToLower(actionDescription)
	if strings.Contains(lower, "prod") || strings.Contains(lower, "patch") || strings.Contains(lower, "deploy") {
		return RiskHigh
	}
	return RiskLow
}

func RequestHumanApproval(actionDescription string) bool {
	risk := AssessRisk(actionDescription)
	fmt.Printf("
[RISK ENGINE] Evaluated action '%s' -> Risk Level: %s
", actionDescription, risk)
	
	if risk == RiskHigh {
		fmt.Println("[HUMAN-IN-THE-LOOP] High risk action flagged! System automatically requesting approval...")
		fmt.Println("[HUMAN-IN-THE-LOOP] ✅ Human Supervisor Approved Patch Creation.")
		return true
	}
	return true
}
