package formatter

import (
	"reflect"
	"testing"
)

func TestSortOrder_alpha(t *testing.T) {
	order := []string{"E501", "B904", "F401"}
	counts := map[string]int{"E501": 3, "B904": 10, "F401": 1}
	got := SortOrder(order, counts, "alpha")
	want := []string{"B904", "E501", "F401"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("alpha sort: got %v, want %v", got, want)
	}
}

func TestSortOrder_count(t *testing.T) {
	order := []string{"E501", "B904", "F401"}
	counts := map[string]int{"E501": 3, "B904": 10, "F401": 1}
	got := SortOrder(order, counts, "count")
	want := []string{"B904", "E501", "F401"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("count sort: got %v, want %v", got, want)
	}
}

func TestSortOrder_countTiebreakAlpha(t *testing.T) {
	order := []string{"Z", "A"}
	counts := map[string]int{"Z": 5, "A": 5}
	got := SortOrder(order, counts, "count")
	if got[0] != "A" {
		t.Errorf("count sort tiebreak: want A first, got %v", got)
	}
}

func TestFilterRuleOrder_empty(t *testing.T) {
	order := []string{"E501", "F401", "B904"}
	got := FilterRuleOrder(order, nil)
	if !reflect.DeepEqual(got, order) {
		t.Error("nil filter should return unchanged")
	}
}

func TestFilterRuleOrder_filters(t *testing.T) {
	order := []string{"E501", "F401", "B904"}
	got := FilterRuleOrder(order, []string{"E501", "B904"})
	want := []string{"E501", "B904"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMatchesFileFilter_empty(t *testing.T) {
	if !MatchesFileFilter("src/foo.py", nil) {
		t.Error("nil filter should match everything")
	}
}

func TestMatchesFileFilter_prefix(t *testing.T) {
	if !MatchesFileFilter("src/foo.py", []string{"src/"}) {
		t.Error("should match src/ prefix")
	}
	if MatchesFileFilter("tests/bar.py", []string{"src/"}) {
		t.Error("should not match tests/ against src/ filter")
	}
}
