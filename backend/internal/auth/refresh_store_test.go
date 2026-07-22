//go:build test

package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/auth"
	authjwt "github.com/JosemaPereira/TrAIveler/backend/internal/auth/jwt"
	authmocks "github.com/JosemaPereira/TrAIveler/backend/internal/auth/mocks"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

func TestUnitRefreshStoreByHash(t *testing.T) {
	ctx := context.Background()
	const hash = "abc123hash"

	t.Run("when the repository returns a live token", func(t *testing.T) {
		t.Run("should map the row into a jwt RefreshTokenRecord", func(t *testing.T) {
			// Arrange
			id := uuid.New()
			userID := uuid.New()
			expiresAt := time.Now().Add(time.Hour).UTC()
			repo := authmocks.NewMockRefreshTokenRepository(t)
			repo.EXPECT().GetRefreshTokenByHash(ctx, hash).Return(&auth.RefreshToken{
				ID:        id.String(),
				UserID:    userID.String(),
				TokenHash: hash,
				ExpiresAt: expiresAt,
			}, nil).Once()

			// Act
			record, err := auth.NewRefreshStore(repo).ByHash(ctx, hash)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, id, record.ID)
			assert.Equal(t, userID, record.UserID)
			assert.Equal(t, expiresAt, record.ExpiresAt)
			assert.Nil(t, record.RevokedAt)
		})
	})

	t.Run("when the repository reports the token is missing", func(t *testing.T) {
		t.Run("should translate domain NotFound into jwt.ErrRefreshTokenNotFound", func(t *testing.T) {
			// Arrange
			repo := authmocks.NewMockRefreshTokenRepository(t)
			repo.EXPECT().GetRefreshTokenByHash(ctx, hash).
				Return(nil, domainerrors.NotFound("refresh token", hash)).Once()

			// Act
			_, err := auth.NewRefreshStore(repo).ByHash(ctx, hash)

			// Assert
			require.ErrorIs(t, err, authjwt.ErrRefreshTokenNotFound)
		})
	})

	t.Run("when the repository returns an unexpected error", func(t *testing.T) {
		t.Run("should propagate the error unchanged", func(t *testing.T) {
			// Arrange
			wantErr := errors.New("connection refused")
			repo := authmocks.NewMockRefreshTokenRepository(t)
			repo.EXPECT().GetRefreshTokenByHash(ctx, hash).Return(nil, wantErr).Once()

			// Act
			_, err := auth.NewRefreshStore(repo).ByHash(ctx, hash)

			// Assert
			require.ErrorIs(t, err, wantErr)
			assert.NotErrorIs(t, err, authjwt.ErrRefreshTokenNotFound)
		})
	})
}

func TestUnitRefreshStoreRevoke(t *testing.T) {
	ctx := context.Background()

	t.Run("when revoking by UUID", func(t *testing.T) {
		t.Run("should delegate to the repository using the string form of the id", func(t *testing.T) {
			// Arrange
			id := uuid.New()
			repo := authmocks.NewMockRefreshTokenRepository(t)
			repo.EXPECT().RevokeRefreshToken(ctx, id.String()).Return(nil).Once()

			// Act
			err := auth.NewRefreshStore(repo).Revoke(ctx, id)

			// Assert
			require.NoError(t, err)
		})
	})
}

func TestUnitRefreshStoreCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("when persisting a newly issued token", func(t *testing.T) {
		t.Run("should map the jwt payload and generate a valid row id", func(t *testing.T) {
			// Arrange
			userID := uuid.New()
			expiresAt := time.Now().Add(30 * 24 * time.Hour).UTC()
			repo := authmocks.NewMockRefreshTokenRepository(t)
			repo.EXPECT().CreateRefreshToken(ctx, mock.MatchedBy(func(tok *auth.RefreshToken) bool {
				_, idErr := uuid.Parse(tok.ID)
				return idErr == nil &&
					tok.UserID == userID.String() &&
					tok.TokenHash == "hashed-token" &&
					tok.ExpiresAt.Equal(expiresAt)
			})).Return(nil).Once()

			// Act
			err := auth.NewRefreshStore(repo).Create(ctx, authjwt.NewRefreshToken{
				UserID:    userID,
				TokenHash: "hashed-token",
				ExpiresAt: expiresAt,
			})

			// Assert
			require.NoError(t, err)
		})
	})
}
