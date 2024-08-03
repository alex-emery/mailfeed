package database

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMigration(t *testing.T) {
	_, err := New(zap.NewNop(), ":memory:")
	require.NoError(t, err)
}
