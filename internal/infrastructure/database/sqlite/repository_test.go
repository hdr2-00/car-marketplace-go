package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"car-marketplace/internal/domain"
	"car-marketplace/migrations"
)

func migratedDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := NewDatabase(*openTestDB(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	if err := Migrate(ctx, db, migrations.Schema); err != nil {
		t.Fatal(err)
	}
	if err := Seed(ctx, db, migrations.Seeds); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestConsumeRefreshTokenSucceedsOnlyOnceUnderConcurrency(t *testing.T) {
	db := migratedDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	if err := repo.CreateRefreshToken(ctx, &domain.RefreshToken{
		UserID: 1, Token: "raw-token", ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}); err != nil {
		t.Fatal(err)
	}

	const attempts = 8
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := repo.ConsumeRefreshToken(ctx, "raw-token"); err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			} else if !errors.Is(err, domain.ErrInvalidRefreshToken) {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if successes != 1 {
		t.Fatalf("token consumed %d times, want exactly 1", successes)
	}
}

func TestConsumeRefreshTokenRejectsExpired(t *testing.T) {
	db := migratedDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	repo.CreateRefreshToken(ctx, &domain.RefreshToken{
		UserID: 1, Token: "old", ExpiresAt: time.Now().Add(-time.Minute), CreatedAt: time.Now(),
	})
	if _, err := repo.ConsumeRefreshToken(ctx, "old"); !errors.Is(err, domain.ErrInvalidRefreshToken) {
		t.Fatalf("expired token: got %v, want ErrInvalidRefreshToken", err)
	}
}

func TestDeleteExpiredTokensRemovesOnlyDeadTokens(t *testing.T) {
	db := migratedDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	repo.CreateRefreshToken(ctx, &domain.RefreshToken{UserID: 1, Token: "expired", ExpiresAt: time.Now().Add(-time.Hour), CreatedAt: time.Now()})
	repo.CreateRefreshToken(ctx, &domain.RefreshToken{UserID: 1, Token: "live", ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now()})

	if err := repo.DeleteExpiredTokens(ctx); err != nil {
		t.Fatal(err)
	}

	var n int
	db.QueryRow(`SELECT COUNT(*) FROM refresh_tokens`).Scan(&n)
	if n != 1 {
		t.Fatalf("tokens left = %d, want 1", n)
	}
	if _, err := repo.ConsumeRefreshToken(ctx, "live"); err != nil {
		t.Fatalf("live token was removed: %v", err)
	}
}

func TestWithinTxRollsBackOnError(t *testing.T) {
	db := migratedDB(t)
	tx := NewTransactor(db)
	users := NewUserRepository(db)
	ctx := context.Background()

	boom := errors.New("boom")
	err := tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := users.Create(ctx, &domain.User{
			Email: "rollback@example.com", PasswordHash: "x", UserTypeID: 1,
			IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("WithinTx error = %v, want boom", err)
	}

	if _, err := users.GetByEmail(ctx, "rollback@example.com"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("user should have been rolled back, got %v", err)
	}
}

func TestDatabaseRejectsDuplicateGeneralReview(t *testing.T) {
	db := migratedDB(t)
	repo := NewReviewRepository(db)
	ctx := context.Background()

	newReview := func() *domain.Review {
		return &domain.Review{ReviewerID: 1, ReviewedUserID: 2, Rating: 4, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	}
	// Remove any seeded general review between these users first.
	db.Exec(`DELETE FROM reviews WHERE reviewer_id = 1 AND reviewed_user_id = 2`)

	if err := repo.Create(ctx, newReview()); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, newReview()); !errors.Is(err, domain.ErrReviewAlreadyExists) {
		t.Fatalf("second general review: got %v, want ErrReviewAlreadyExists", err)
	}
}
