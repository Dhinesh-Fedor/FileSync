package report

import (
	"encoding/json"
	"fmt"
	"os"

	"FileSync/internal/models"
)

func Save(pairID int, report models.SyncReport) error {

	err := os.MkdirAll("reports", 0755)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("reports/pair-%d.json", pairID)

	return os.WriteFile(fileName, data, 0644)
}
