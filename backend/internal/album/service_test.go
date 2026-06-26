package album

import "testing"

func TestNormalizeRejectsEmptyAlbumName(t *testing.T) {
	_, err := normalize(Input{Name: " ", Value: "travel"})
	if err != ErrInvalidName {
		t.Fatalf("normalize() error = %v, want ErrInvalidName", err)
	}
}
