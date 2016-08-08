package models

import (
	"fmt"
	"sort"
	"testing"
	"time"
)

func Test_playlistTitleRegexpMatch(t *testing.T) {
	type testCases struct {
		candidate string
		result    bool
	}

	cases := []testCases{
		{
			candidate: "Watch Later Service",
			result:    true,
		},
		{
			candidate: "Watch Later Service 2",
			result:    true,
		},
		{
			candidate: "Watch Later Service 3",
			result:    true,
		},
		{
			candidate: "Watch Later Service 15",
			result:    true,
		},
		{
			candidate: "Watch Later Service: my favorite",
			result:    false,
		},
		{
			candidate: "Watch Later Service 1-3 minutes",
			result:    false,
		},
	}

	num_cases := len(cases)
	for i, c := range cases {
		case_index := i + 1

		result := playlistTitleRegexpMatch(c.candidate)
		if result != c.result {
			t.Errorf("\n[%d of %d: Results should equal] \n\t%v \n \n\t%v", case_index, num_cases, result, c.result)
		}
	}
}

func Test_extractISO8601Duration(t *testing.T) {
	type testCases struct {
		duration string
		minutes  time.Duration
		seconds  time.Duration
		err      error
	}

	cases := []testCases{
		{
			duration: "PT15M51S",
			minutes:  time.Minute * 15,
			seconds:  time.Second * 51,
			err:      nil,
		},
		{
			duration: "PT3M45S",
			minutes:  time.Minute * 3,
			seconds:  time.Second * 45,
			err:      nil,
		},
		{
			duration: "PT150M510S",
			minutes:  time.Minute * 150,
			seconds:  time.Second * 510,
			err:      nil,
		},
		{
			duration: "PT15M51Second",
			minutes:  0,
			seconds:  0,
			err:      fmt.Errorf(`No duration can be extracted`),
		},
	}

	num_cases := len(cases)
	for i, c := range cases {
		case_index := i + 1

		m, s, err := extractISO8601Duration(c.duration)
		if m != c.minutes {
			t.Errorf("\n[%d of %d: Results should equal] \n\t%v \n \n\t%v", case_index, num_cases, m, c.minutes)
		}
		if s != c.seconds {
			t.Errorf("\n[%d of %d: Results should equal] \n\t%v \n \n\t%v", case_index, num_cases, s, c.seconds)
		}

		var (
			original_error_string,
			test_case_error_string string
		)

		if err != nil {
			original_error_string = err.Error()
		}
		if c.err != nil {
			test_case_error_string = c.err.Error()
		}

		if original_error_string != test_case_error_string {
			t.Errorf("\n[%d of %d: Errors should equal] \n\t%v \n \n\t%v", case_index, num_cases, err, c.err)
		}
	}
}

func Test_checkVideoDuration(t *testing.T) {
	type testCases struct {
		duration          time.Duration
		duration_playlist *DurationPlaylist
		result            bool
	}

	cases := []testCases{
		{
			duration: time.Minute,
			duration_playlist: &DurationPlaylist{
				start: time.Second,
				end:   time.Minute*3 + time.Second*59,
			},
			result: true,
		},
		{
			duration: time.Second,
			duration_playlist: &DurationPlaylist{
				start: time.Second,
				end:   time.Minute*3 + time.Second*59,
			},
			result: true,
		},
		{
			duration: time.Minute * 3,
			duration_playlist: &DurationPlaylist{
				start: time.Second,
				end:   time.Minute*3 + time.Second*59,
			},
			result: true,
		},
		{
			duration: time.Minute*3 + time.Second*59,
			duration_playlist: &DurationPlaylist{
				start: time.Second,
				end:   time.Minute*3 + time.Second*59,
			},
			result: true,
		},
		{
			duration: time.Minute * 4,
			duration_playlist: &DurationPlaylist{
				start: time.Second,
				end:   time.Minute*3 + time.Second*59,
			},
			result: false,
		},
		///
		{
			duration: time.Minute * 2,
			duration_playlist: &DurationPlaylist{
				start: time.Minute * 3,
				end:   time.Minute*3 + time.Second*59,
			},
			result: false,
		},
	}

	num_cases := len(cases)
	for i, c := range cases {
		case_index := i + 1

		result := c.duration_playlist.checkMatches(c.duration)
		if result != c.result {
			t.Errorf("\n[%d of %d: Results should equal] \n\t%v \n \n\t%v", case_index, num_cases, result, c.result)
		}
	}

}

func Test_VideosAreSortable(t *testing.T) {
	videos := Videos{
		&Video{
			Duration: time.Minute + time.Second*11,
		},
		&Video{
			Duration: time.Minute + time.Second*10,
		},
		&Video{
			Duration: time.Minute + time.Second*13,
		},
		&Video{
			Duration: time.Minute + time.Second*12,
		},
		&Video{
			Duration: time.Minute + time.Second*16,
		},
	}

	sort.Sort(videos)

	if videos[0].Duration != time.Minute+time.Second*10 {
		t.Errorf("Sort failed")
	}
	if videos[1].Duration != time.Minute+time.Second*11 {
		t.Errorf("Sort failed")
	}
	if videos[2].Duration != time.Minute+time.Second*12 {
		t.Errorf("Sort failed")
	}
	if videos[3].Duration != time.Minute+time.Second*13 {
		t.Errorf("Sort failed")
	}
	if videos[4].Duration != time.Minute+time.Second*16 {
		t.Errorf("Sort failed")
	}
}

func Test_PlaylistsAreSortable(t *testing.T) {
	playlists := Playlists{
		&Playlist{
			Title: "Watch Later Service 5",
		},
		&Playlist{
			Title: "Watch Later Service 3",
		},
		&Playlist{
			Title: "Watch Later Service",
		},
		&Playlist{
			Title: "Watch Later Service 2",
		},
		&Playlist{
			Title: "Watch Later Service 3",
		},
	}

	sort.Sort(playlists)

	if playlists[0].Title != `Watch Later Service` {
		t.Errorf("Sort failed")
	}
	if playlists[1].Title != `Watch Later Service 2` {
		t.Errorf("Sort failed")
	}
	if playlists[2].Title != `Watch Later Service 3` {
		t.Errorf("Sort failed")
	}
	if playlists[3].Title != `Watch Later Service 3` {
		t.Errorf("Sort failed")
	}
	if playlists[4].Title != `Watch Later Service 5` {
		t.Errorf("Sort failed")
	}
}

func Test_compareWatchLaterPlaylistTitles(t *testing.T) {
	type testCases struct {
		a      string
		b      string
		result bool
	}

	cases := []testCases{
		{
			a:      "Watch Later Service 2",
			b:      "Watch Later Service",
			result: true,
		},
		{
			a:      "Watch Later Service",
			b:      "Watch Later Service 2",
			result: false,
		},
		{
			a:      "Watch Later Service",
			b:      "Watch Later Service",
			result: false,
		},
	}

	num_cases := len(cases)
	for i, c := range cases {
		case_index := i + 1

		result := compareWatchLaterPlaylistTitles(c.a, c.b)
		if result != c.result {
			t.Errorf("\n[%d of %d: Results should equal] \n\t%v \n \n\t%v", case_index, num_cases, result, c.result)
		}
	}
}

func Test_getNextWatchLaterServicePlaylistDigit(t *testing.T) {
	type testCases struct {
		candidate string
		result    int64
		err       error
	}

	cases := []testCases{
		{
			candidate: "",
			result:    -1,
			err:       fmt.Errorf("Need a name of a playlist"),
		},
		{
			candidate: "Watch Later Service",
			result:    2,
			err:       nil,
		},
		{
			candidate: "Watch Later Service 2",
			result:    3,
			err:       nil,
		},
		{
			candidate: "Watch Later Service 15",
			result:    16,
			err:       nil,
		},
		{
			candidate: "Watch Later Service: my favorite",
			result:    -1,
			err:       fmt.Errorf(`strconv.ParseInt: parsing ": my favorite": invalid syntax`),
		},
		{
			candidate: "Watch Later Service 1-3 minutes",
			result:    -1,
			err:       fmt.Errorf(`strconv.ParseInt: parsing "1-3 minutes": invalid syntax`),
		},
	}

	num_cases := len(cases)
	for i, c := range cases {
		case_index := i + 1

		result, err := getNextWatchLaterServicePlaylistDigit(c.candidate)
		if result != c.result {
			t.Errorf("\n[%d of %d: Results should equal] \n\t%v \n \n\t%v", case_index, num_cases, result, c.result)
		}

		var (
			original_error_string,
			test_case_error_string string
		)

		if err != nil {
			original_error_string = err.Error()
		}
		if c.err != nil {
			test_case_error_string = c.err.Error()
		}

		if original_error_string != test_case_error_string {
			t.Errorf("\n[%d of %d: Errors should equal] \n\t%v \n \n\t%v", case_index, num_cases, err, c.err)
		}
	}
}
