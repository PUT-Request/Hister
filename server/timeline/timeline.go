// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package timeline

import "time"

const recentDayCount = 7

type Bucket struct {
	Key   string `json:"key"`
	From  int64  `json:"from,omitempty"`
	To    int64  `json:"to"`
	Count int    `json:"count"`
}

type BucketList []Bucket

type Result struct {
	Days   BucketList `json:"days"`
	Older  Bucket     `json:"older"`
	Months BucketList `json:"months"`
}

type DailyResult struct {
	Days BucketList `json:"days"`
}

func New(now time.Time, loc *time.Location, oldest int64) *Result {
	if loc == nil {
		loc = time.UTC
	}
	now = now.In(loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	result := &Result{
		Days:   make(BucketList, 0, recentDayCount),
		Months: BucketList{},
	}

	for day := range recentDayCount {
		from := today.AddDate(0, 0, -day)
		to := from.AddDate(0, 0, 1)
		result.Days = append(result.Days, Bucket{
			Key:  "day:" + from.Format(time.DateOnly),
			From: from.Unix(),
			To:   to.Unix(),
		})
	}

	olderBefore := today.AddDate(0, 0, -(recentDayCount - 1))
	result.Older = Bucket{Key: "older", To: olderBefore.Unix()}
	result.Months = monthBuckets(olderBefore, oldest)
	return result
}

func NewDays(dateFrom, dateTo int64, loc *time.Location) *DailyResult {
	result := &DailyResult{Days: BucketList{}}
	if dateFrom >= dateTo {
		return result
	}
	if loc == nil {
		loc = time.UTC
	}

	from := time.Unix(dateFrom, 0).In(loc)
	day := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, loc)
	for day.Unix() < dateTo {
		nextDay := day.AddDate(0, 0, 1)
		bucketFrom := max(day.Unix(), dateFrom)
		bucketTo := min(nextDay.Unix(), dateTo)
		if bucketFrom < bucketTo {
			result.Days = append(result.Days, Bucket{
				Key:  "day:" + day.Format(time.DateOnly),
				From: bucketFrom,
				To:   bucketTo,
			})
		}
		day = nextDay
	}
	for left, right := 0, len(result.Days)-1; left < right; left, right = left+1, right-1 {
		result.Days[left], result.Days[right] = result.Days[right], result.Days[left]
	}
	return result
}

func monthBuckets(before time.Time, oldest int64) BucketList {
	if oldest == 0 || oldest >= before.Unix() {
		return BucketList{}
	}
	months := make(BucketList, 0)
	cursor := before
	monthStart := time.Date(cursor.Year(), cursor.Month(), 1, 0, 0, 0, 0, cursor.Location())
	if monthStart.Before(cursor) {
		months = append(months, Bucket{
			Key:  "month:" + monthStart.Format("2006-01"),
			From: monthStart.Unix(),
			To:   cursor.Unix(),
		})
		cursor = monthStart
	}
	for oldest < cursor.Unix() {
		previous := cursor.AddDate(0, -1, 0)
		months = append(months, Bucket{
			Key:  "month:" + previous.Format("2006-01"),
			From: previous.Unix(),
			To:   cursor.Unix(),
		})
		cursor = previous
	}
	return months
}

func (r *Result) Buckets() []*Bucket {
	buckets := make([]*Bucket, 0, len(r.Days)+len(r.Months)+1)
	buckets = append(buckets, r.Days.Pointers()...)
	buckets = append(buckets, &r.Older)
	buckets = append(buckets, r.Months.Pointers()...)
	return buckets
}

func (r *Result) SetCount(key string, count int) {
	if r.Days.SetCount(key, count) {
		return
	}
	if r.Older.Key == key {
		r.Older.Count = count
		return
	}
	r.Months.SetCount(key, count)
}

func (r *Result) AddTimestamp(timestamp int64) {
	if r.Days.AddTimestamp(timestamp) {
		return
	}
	if !r.Older.contains(timestamp) {
		return
	}
	r.Older.Count++
	r.Months.AddTimestamp(timestamp)
}

func (r *DailyResult) Buckets() []*Bucket {
	return r.Days.Pointers()
}

func (r *DailyResult) SetCount(key string, count int) {
	r.Days.SetCount(key, count)
}

func (r *DailyResult) AddTimestamp(timestamp int64) {
	r.Days.AddTimestamp(timestamp)
}

func (buckets BucketList) Pointers() []*Bucket {
	result := make([]*Bucket, 0, len(buckets))
	for i := range buckets {
		result = append(result, &buckets[i])
	}
	return result
}

func (buckets BucketList) SetCount(key string, count int) bool {
	for i := range buckets {
		if buckets[i].Key == key {
			buckets[i].Count = count
			return true
		}
	}
	return false
}

func (buckets BucketList) AddTimestamp(timestamp int64) bool {
	for i := range buckets {
		if buckets[i].contains(timestamp) {
			buckets[i].Count++
			return true
		}
	}
	return false
}

func (b Bucket) contains(timestamp int64) bool {
	return (b.From == 0 || timestamp >= b.From) && timestamp < b.To
}
