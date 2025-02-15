package light

import (
	"context"
	_ "embed"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/celestiaorg/celestia-app/pkg/da"
	"github.com/celestiaorg/celestia-app/pkg/namespace"

	"github.com/celestiaorg/celestia-node/header"
	"github.com/celestiaorg/celestia-node/share"
	availability_test "github.com/celestiaorg/celestia-node/share/availability/test"
)

func TestSharesAvailable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getter, dah := GetterWithRandSquare(t, 16)
	avail := TestAvailability(getter)
	err := avail.SharesAvailable(ctx, dah)
	assert.NoError(t, err)
}

func TestSharesAvailableFailed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getter, _ := GetterWithRandSquare(t, 16)
	avail := TestAvailability(getter)
	empty := header.EmptyDAH()
	err := avail.SharesAvailable(ctx, &empty)
	assert.Error(t, err)
}

func TestShareAvailableOverMocknet_Light(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	net := availability_test.NewTestDAGNet(ctx, t)
	_, root := RandNode(net, 16)
	nd := Node(net)
	net.ConnectAll()

	err := nd.SharesAvailable(ctx, root)
	assert.NoError(t, err)
}

func TestGetShare(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := 16
	getter, dah := GetterWithRandSquare(t, n)

	for i := range make([]bool, n) {
		for j := range make([]bool, n) {
			sh, err := getter.GetShare(ctx, dah, i, j)
			assert.NotNil(t, sh)
			assert.NoError(t, err)
		}
	}
}

func TestService_GetSharesByNamespace(t *testing.T) {
	var tests = []struct {
		squareSize         int
		expectedShareCount int
	}{
		{squareSize: 4, expectedShareCount: 2},
		{squareSize: 16, expectedShareCount: 2},
		{squareSize: 128, expectedShareCount: 2},
	}

	for _, tt := range tests {
		t.Run("size: "+strconv.Itoa(tt.squareSize), func(t *testing.T) {
			getter, bServ := EmptyGetter()
			totalShares := tt.squareSize * tt.squareSize
			randShares := share.RandShares(t, totalShares)
			idx1 := (totalShares - 1) / 2
			idx2 := totalShares / 2
			if tt.expectedShareCount > 1 {
				// make it so that two rows have the same namespace ID
				copy(randShares[idx2][:namespace.NamespaceSize], randShares[idx1][:namespace.NamespaceSize])
			}
			root := availability_test.FillBS(t, bServ, randShares)
			randNID := randShares[idx1][:namespace.NamespaceSize]

			shares, err := getter.GetSharesByNamespace(context.Background(), root, randNID)
			require.NoError(t, err)
			require.NoError(t, shares.Verify(root, randNID))
			flattened := shares.Flatten()
			assert.Len(t, flattened, tt.expectedShareCount)
			for _, value := range flattened {
				assert.Equal(t, randNID, []byte(share.ID(value)))
			}
			if tt.expectedShareCount > 1 {
				// idx1 is always smaller than idx2
				assert.Equal(t, randShares[idx1], flattened[0])
				assert.Equal(t, randShares[idx2], flattened[1])
			}
		})
		t.Run("last two rows of a 4x4 square that have the same namespace ID have valid NMT proofs", func(t *testing.T) {
			squareSize := 4
			totalShares := squareSize * squareSize
			getter, bServ := EmptyGetter()
			randShares := share.RandShares(t, totalShares)
			lastNID := randShares[totalShares-1][:namespace.NamespaceSize]
			for i := totalShares / 2; i < totalShares; i++ {
				copy(randShares[i][:namespace.NamespaceSize], lastNID)
			}
			root := availability_test.FillBS(t, bServ, randShares)

			shares, err := getter.GetSharesByNamespace(context.Background(), root, lastNID)
			require.NoError(t, err)
			require.NoError(t, shares.Verify(root, lastNID))
		})
	}
}

func TestGetShares(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := 16
	getter, dah := GetterWithRandSquare(t, n)

	eds, err := getter.GetEDS(ctx, dah)
	require.NoError(t, err)
	gotDAH := da.NewDataAvailabilityHeader(eds)

	require.True(t, dah.Equals(&gotDAH))
}

func TestService_GetSharesByNamespaceNotFound(t *testing.T) {
	getter, root := GetterWithRandSquare(t, 1)
	root.RowRoots = nil

	_, err := getter.GetSharesByNamespace(context.Background(), root, namespace.RandomNamespace().Bytes())
	assert.ErrorIs(t, err, share.ErrNamespaceNotFound)
}

func BenchmarkService_GetSharesByNamespace(b *testing.B) {
	var tests = []struct {
		amountShares int
	}{
		{amountShares: 4},
		{amountShares: 16},
		{amountShares: 128},
	}

	for _, tt := range tests {
		b.Run(strconv.Itoa(tt.amountShares), func(b *testing.B) {
			t := &testing.T{}
			getter, root := GetterWithRandSquare(t, tt.amountShares)
			randNID := root.RowRoots[(len(root.RowRoots)-1)/2][:8]
			root.RowRoots[(len(root.RowRoots) / 2)] = root.RowRoots[(len(root.RowRoots)-1)/2]
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := getter.GetSharesByNamespace(context.Background(), root, randNID)
				require.NoError(t, err)
			}
		})
	}
}
