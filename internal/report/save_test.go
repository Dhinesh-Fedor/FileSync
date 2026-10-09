package report

import (
	"os"
	"testing"
	"time"

	"FileSync/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	root := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	report := models.SyncReport{
		Added:       3,
		Modified:    2,
		Deleted:     1,
		BytesCopied: 1024,
		StartTime:   time.Now(),
		EndTime:     time.Now(),
	}

	err = Save(1, report)

	assert.NoError(t, err)

	_, err = os.Stat("storage/client/logs/pair-1.json")

	assert.NoError(t, err)

	_ = os.Remove("storage/client/logs/pair-1.json")
}
