package service

import (
	"reflect"
	"testing"
)

func TestExtractTags_Basic(t *testing.T) {
	tags := ExtractTags("Hello #world #go")
	want := []string{"world", "go"}
	if !reflect.DeepEqual(tags, want) {
		t.Errorf("got %v, want %v", tags, want)
	}
}

func TestExtractTags_Dedup(t *testing.T) {
	tags := ExtractTags("#go is great #go")
	want := []string{"go"}
	if !reflect.DeepEqual(tags, want) {
		t.Errorf("got %v, want %v", tags, want)
	}
}

func TestExtractTags_Japanese(t *testing.T) {
	tags := ExtractTags("これは #日本語タグ です")
	want := []string{"日本語タグ"}
	if !reflect.DeepEqual(tags, want) {
		t.Errorf("got %v, want %v", tags, want)
	}
}

func TestExtractTags_NoTags(t *testing.T) {
	tags := ExtractTags("Hello world, no tags here")
	if len(tags) != 0 {
		t.Errorf("expected empty slice, got %v", tags)
	}
}

func TestExtractTags_AtStartOfLine(t *testing.T) {
	tags := ExtractTags("#first tag")
	want := []string{"first"}
	if !reflect.DeepEqual(tags, want) {
		t.Errorf("got %v, want %v", tags, want)
	}
}

func TestExtractTags_WithUnderscore(t *testing.T) {
	tags := ExtractTags("#my_tag")
	want := []string{"my_tag"}
	if !reflect.DeepEqual(tags, want) {
		t.Errorf("got %v, want %v", tags, want)
	}
}

func TestExtractTags_WithNumbers(t *testing.T) {
	tags := ExtractTags("#tag123")
	want := []string{"tag123"}
	if !reflect.DeepEqual(tags, want) {
		t.Errorf("got %v, want %v", tags, want)
	}
}
