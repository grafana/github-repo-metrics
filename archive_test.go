package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var startDate = time.Date(2018, time.Month(1), 1, 1, 0, 0, 0, time.Local)

func TestGenerateUrlsFor6Hours(t *testing.T) {
	var archivedFiles []*ArchiveFile

	stopDate := fakeStopDate(1, 6)
	result := buildUrlsDownload(archivedFiles, startDate, stopDate)

	assert.Equal(t, len(result), 6, "failure")
}

func TestGenerateUrlsForOneDay(t *testing.T) {
	var archivedFiles []*ArchiveFile

	stopDate := fakeStopDate(2, 0)

	result := buildUrlsDownload(archivedFiles, startDate, stopDate)

	assert.Equal(t, len(result), 24, "failure")
}

func TestGenerateUrlsForTwoAndAHalfDay(t *testing.T) {
	var archivedFiles []*ArchiveFile

	stopDate := fakeStopDate(2, 12)

	result := buildUrlsDownload(archivedFiles, startDate, stopDate)

	assert.Equal(t, len(result), 36, "failure")
}

func TestGenerateUrlsWhenArchivedFilesExists(t *testing.T) {
	var archivedFiles []*ArchiveFile

	for i := 0; i < 12; i++ {
		archivedFiles = append(archivedFiles, &ArchiveFile{
			Year:  2018,
			Month: 1,
			Day:   1,
			Hour:  i})
	}

	stopDate := fakeStopDate(1, 0)

	result := buildUrlsDownload(archivedFiles, startDate, stopDate)

	assert.Equal(t, len(result), 0, "failure")
}

func fakeStopDate(days, hours int) time.Time {
	return time.Date(2018, time.Month(1), days, hours, 0, 0, 0, time.Local)
}