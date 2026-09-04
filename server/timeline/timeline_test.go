// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package timeline

import (
	"testing"
	"time"
)

func TestNewBuildsConsecutiveBuckets(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Budapest")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.July, 28, 15, 0, 0, 0, loc)
	oldest := time.Date(2026, time.March, 10, 12, 0, 0, 0, loc).Unix()
	result := New(now, loc, oldest)

	if len(result.Days) != 7 {
		t.Fatalf("daily bucket count = %d, want 7", len(result.Days))
	}
	if result.Days[len(result.Days)-1].From != result.Older.To {
		t.Fatal("daily and older buckets are not consecutive")
	}
	if len(result.Months) == 0 || result.Months[0].To != result.Older.To {
		t.Fatal("monthly buckets do not start at the older boundary")
	}
	for i := 1; i < len(result.Months); i++ {
		if result.Months[i].To != result.Months[i-1].From {
			t.Fatalf("monthly buckets %d and %d are not consecutive", i-1, i)
		}
	}
}

func TestNewUsesCalendarDaysAcrossDaylightSaving(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Budapest")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, time.March, 30, 12, 0, 0, 0, loc)
	result := New(now, loc, 0)

	var foundShortDay bool
	for _, bucket := range result.Days {
		if bucket.To-bucket.From == 23*time.Hour.Milliseconds()/1000 {
			foundShortDay = true
		}
	}
	if !foundShortDay {
		t.Fatal("daily buckets did not preserve the daylight saving boundary")
	}
}

func TestNewDaysBuildsDescendingCalendarBuckets(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Budapest")
	if err != nil {
		t.Fatal(err)
	}
	dateFrom := time.Date(2026, time.March, 28, 0, 0, 0, 0, loc).Unix()
	dateTo := time.Date(2026, time.April, 2, 0, 0, 0, 0, loc).Unix()
	result := NewDays(dateFrom, dateTo, loc)

	if len(result.Days) != 5 {
		t.Fatalf("daily bucket count = %d, want 5", len(result.Days))
	}
	if result.Days[0].Key != "day:2026-04-01" || result.Days[4].Key != "day:2026-03-28" {
		t.Fatalf("daily bucket order = %q to %q", result.Days[0].Key, result.Days[4].Key)
	}
	for i := 1; i < len(result.Days); i++ {
		if result.Days[i-1].From != result.Days[i].To {
			t.Fatalf("daily buckets %d and %d are not consecutive", i-1, i)
		}
	}
	result.AddTimestamp(time.Date(2026, time.March, 29, 12, 0, 0, 0, loc).Unix())
	if result.Days[3].Count != 1 {
		t.Fatalf("March 29 count = %d, want 1", result.Days[3].Count)
	}
}

func TestAddTimestampCountsOneVisiblePeriodAndOlderMonth(t *testing.T) {
	loc := time.UTC
	now := time.Date(2026, time.July, 28, 12, 0, 0, 0, loc)
	oldest := time.Date(2026, time.April, 1, 0, 0, 0, 0, loc).Unix()
	result := New(now, loc, oldest)

	result.AddTimestamp(time.Date(2026, time.July, 28, 1, 0, 0, 0, loc).Unix())
	result.AddTimestamp(time.Date(2026, time.July, 10, 1, 0, 0, 0, loc).Unix())
	result.AddTimestamp(time.Date(2026, time.April, 10, 1, 0, 0, 0, loc).Unix())

	if result.Days[0].Count != 1 {
		t.Fatalf("today count = %d, want 1", result.Days[0].Count)
	}
	if result.Older.Count != 2 {
		t.Fatalf("older count = %d, want 2", result.Older.Count)
	}
	var julyCount, aprilCount int
	for _, bucket := range result.Months {
		switch bucket.Key {
		case "month:2026-07":
			julyCount = bucket.Count
		case "month:2026-04":
			aprilCount = bucket.Count
		}
	}
	if julyCount != 1 {
		t.Fatalf("July count = %d, want 1", julyCount)
	}
	if aprilCount != 1 {
		t.Fatalf("April count = %d, want 1", aprilCount)
	}
}
