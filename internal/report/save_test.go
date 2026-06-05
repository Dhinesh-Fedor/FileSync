package report

import (
	"os"
	"testing"
	"time"

	"FileSync/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {

	report := models.SyncReport{
		Added: 3,
		Modified: 2,
		Deleted: 1,
		BytesCopied: 1024,
		StartTime: time.Now(),
		EndTime: time.Now(),
	}

	err := Save(1, report)

	assert.NoError(t, err)

	_, err = os.Stat("reports/pair-1.json")

	assert.NoError(t, err)

	os.Remove("reports/pair-1.json")
}