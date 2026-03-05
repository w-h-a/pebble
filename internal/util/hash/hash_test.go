package hash

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFieldsDeterministic(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	fields := map[string]string{
		"b": "2",
		"a": "1",
		"c": "3",
	}

	// Act
	h1 := Fields(fields)
	h2 := Fields(fields)

	// Assert
	require.Equal(t, h1, h2)
}

func TestFieldsOrderIndependent(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	a := map[string]string{"x": "1", "y": "2"}
	b := map[string]string{"y": "2", "x": "1"}

	// Act
	h1 := Fields(a)
	h2 := Fields(b)

	// Assert
	require.Equal(t, h1, h2)
}

func TestFieldsDifferentValues(t *testing.T) {
	if len(os.Getenv("INTEGRATION")) > 0 {
		t.Skip()
	}

	// Arrange
	a := map[string]string{"key": "val1"}
	b := map[string]string{"key": "val2"}

	// Act
	h1 := Fields(a)
	h2 := Fields(b)

	// Assert
	require.NotEqual(t, h1, h2)
}
