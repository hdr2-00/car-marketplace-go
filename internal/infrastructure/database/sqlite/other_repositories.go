package sqlite

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ================= INQUIRY REPOSITORY =================

type inquiryRepository struct {
	db *sql.DB
}

func NewInquiryRepository(db *sql.DB) repository.InquiryRepository {
	return &inquiryRepository{db: db}
}

func (r *inquiryRepository) Create(ctx context.Context, inquiry *domain.Inquiry) error {
	query := `
		INSERT INTO inquiries (car_listing_id, buyer_id, seller_id, subject, message, is_read, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		inquiry.CarListingID, inquiry.BuyerID, inquiry.SellerID, inquiry.Subject,
		inquiry.Message, inquiry.IsRead, inquiry.CreatedAt,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	inquiry.ID = id
	return nil
}

func (r *inquiryRepository) GetByID(ctx context.Context, id int64) (*domain.Inquiry, error) {
	query := `
		SELECT id, car_listing_id, buyer_id, seller_id, COALESCE(subject, ''), message, is_read, created_at
		FROM inquiries
		WHERE id = ?
	`

	inquiry := &domain.Inquiry{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inquiry.ID, &inquiry.CarListingID, &inquiry.BuyerID, &inquiry.SellerID,
		&inquiry.Subject, &inquiry.Message, &inquiry.IsRead, &inquiry.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInquiryNotFound
		}
		return nil, err
	}

	return inquiry, nil
}

func (r *inquiryRepository) List(ctx context.Context, filter *domain.InquiryFilter) ([]domain.Inquiry, int64, error) {
	where, args := r.buildInquiryWhere(filter)

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM inquiries %s", where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []domain.Inquiry{}, 0, nil
	}

	// Get data
	offset := (filter.Page - 1) * filter.Limit
	query := fmt.Sprintf(`
		SELECT id, car_listing_id, buyer_id, seller_id, COALESCE(subject, ''), message, is_read, created_at
		FROM inquiries
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, where)

	args = append(args, filter.Limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var inquiries []domain.Inquiry
	for rows.Next() {
		var inq domain.Inquiry
		err := rows.Scan(&inq.ID, &inq.CarListingID, &inq.BuyerID, &inq.SellerID,
			&inq.Subject, &inq.Message, &inq.IsRead, &inq.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		inquiries = append(inquiries, inq)
	}

	return inquiries, total, nil
}

func (r *inquiryRepository) MarkAsRead(ctx context.Context, id int64) error {
	query := `UPDATE inquiries SET is_read = 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *inquiryRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM inquiries WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrInquiryNotFound
	}

	return nil
}

func (r *inquiryRepository) buildInquiryWhere(filter *domain.InquiryFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if filter.CarListingID != nil {
		conditions = append(conditions, "car_listing_id = ?")
		args = append(args, *filter.CarListingID)
	}

	if filter.BuyerID != nil {
		conditions = append(conditions, "buyer_id = ?")
		args = append(args, *filter.BuyerID)
	}

	if filter.SellerID != nil {
		conditions = append(conditions, "seller_id = ?")
		args = append(args, *filter.SellerID)
	}

	if filter.IsRead != nil {
		conditions = append(conditions, "is_read = ?")
		args = append(args, *filter.IsRead)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

// ================= REVIEW REPOSITORY =================

type reviewRepository struct {
	db *sql.DB
}

func NewReviewRepository(db *sql.DB) repository.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review *domain.Review) error {
	query := `
		INSERT INTO reviews (reviewer_id, reviewed_user_id, car_listing_id, rating, title, comment, is_verified_purchase, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		review.ReviewerID, review.ReviewedUserID, review.CarListingID, review.Rating,
		review.Title, review.Comment, review.IsVerifiedPurchase, review.CreatedAt, review.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrReviewAlreadyExists
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	review.ID = id
	return nil
}

func (r *reviewRepository) GetByID(ctx context.Context, id int64) (*domain.Review, error) {
	query := `
		SELECT id, reviewer_id, reviewed_user_id, car_listing_id, rating, COALESCE(title, ''), COALESCE(comment, ''), is_verified_purchase, created_at, updated_at
		FROM reviews
		WHERE id = ?
	`

	review := &domain.Review{}
	var carListingID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&review.ID, &review.ReviewerID, &review.ReviewedUserID, &carListingID,
		&review.Rating, &review.Title, &review.Comment, &review.IsVerifiedPurchase,
		&review.CreatedAt, &review.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReviewNotFound
		}
		return nil, err
	}

	if carListingID.Valid {
		review.CarListingID = &carListingID.Int64
	}

	return review, nil
}

func (r *reviewRepository) Update(ctx context.Context, review *domain.Review) error {
	query := `
		UPDATE reviews
		SET rating = ?, title = ?, comment = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		review.Rating, review.Title, review.Comment, review.UpdatedAt, review.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrReviewNotFound
	}

	return nil
}

func (r *reviewRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM reviews WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrReviewNotFound
	}

	return nil
}

func (r *reviewRepository) List(ctx context.Context, filter *domain.ReviewFilter) ([]domain.Review, int64, error) {
	where, args := r.buildReviewWhere(filter)

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM reviews %s", where)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []domain.Review{}, 0, nil
	}

	// Get data
	offset := (filter.Page - 1) * filter.Limit
	query := fmt.Sprintf(`
		SELECT id, reviewer_id, reviewed_user_id, car_listing_id, rating, COALESCE(title, ''), COALESCE(comment, ''), is_verified_purchase, created_at, updated_at
		FROM reviews
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, where)

	args = append(args, filter.Limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var rev domain.Review
		var carListingID sql.NullInt64
		err := rows.Scan(&rev.ID, &rev.ReviewerID, &rev.ReviewedUserID, &carListingID,
			&rev.Rating, &rev.Title, &rev.Comment, &rev.IsVerifiedPurchase,
			&rev.CreatedAt, &rev.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		if carListingID.Valid {
			rev.CarListingID = &carListingID.Int64
		}
		reviews = append(reviews, rev)
	}

	return reviews, total, nil
}

func (r *reviewRepository) GetUserStats(ctx context.Context, userID int64) (*domain.ReviewStats, error) {
	query := `
		SELECT
			COUNT(*) as total_reviews,
			AVG(rating) as average_rating,
			SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END) as rating_1,
			SUM(CASE WHEN rating = 2 THEN 1 ELSE 0 END) as rating_2,
			SUM(CASE WHEN rating = 3 THEN 1 ELSE 0 END) as rating_3,
			SUM(CASE WHEN rating = 4 THEN 1 ELSE 0 END) as rating_4,
			SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END) as rating_5
		FROM reviews
		WHERE reviewed_user_id = ?
	`

	var total int
	var avgRating sql.NullFloat64
	var r1, r2, r3, r4, r5 int

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&total, &avgRating, &r1, &r2, &r3, &r4, &r5,
	)

	if err != nil {
		return nil, err
	}

	stats := &domain.ReviewStats{
		TotalReviews:  total,
		AverageRating: avgRating.Float64,
		RatingCounts: map[int]int{
			1: r1,
			2: r2,
			3: r3,
			4: r4,
			5: r5,
		},
	}

	return stats, nil
}

func (r *reviewRepository) HasReviewed(ctx context.Context, reviewerID, reviewedUserID int64, carListingID *int64) (bool, error) {
	query := `SELECT COUNT(*) FROM reviews WHERE reviewer_id = ? AND reviewed_user_id = ?`
	args := []interface{}{reviewerID, reviewedUserID}

	if carListingID != nil {
		query += " AND car_listing_id = ?"
		args = append(args, *carListingID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *reviewRepository) buildReviewWhere(filter *domain.ReviewFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if filter.ReviewerID != nil {
		conditions = append(conditions, "reviewer_id = ?")
		args = append(args, *filter.ReviewerID)
	}

	if filter.ReviewedUserID != nil {
		conditions = append(conditions, "reviewed_user_id = ?")
		args = append(args, *filter.ReviewedUserID)
	}

	if filter.CarListingID != nil {
		conditions = append(conditions, "car_listing_id = ?")
		args = append(args, *filter.CarListingID)
	}

	if filter.MinRating != nil {
		conditions = append(conditions, "rating >= ?")
		args = append(args, *filter.MinRating)
	}

	if filter.MaxRating != nil {
		conditions = append(conditions, "rating <= ?")
		args = append(args, *filter.MaxRating)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

// ================= FAVORITE REPOSITORY =================

type favoriteRepository struct {
	db *sql.DB
}

func NewFavoriteRepository(db *sql.DB) repository.FavoriteRepository {
	return &favoriteRepository{db: db}
}

func (r *favoriteRepository) Create(ctx context.Context, favorite *domain.Favorite) error {
	query := `
		INSERT INTO favorites (user_id, car_listing_id, created_at)
		VALUES (?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		favorite.UserID, favorite.CarListingID, favorite.CreatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return domain.ErrFavoriteAlreadyExists
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	favorite.ID = id
	return nil
}

func (r *favoriteRepository) Delete(ctx context.Context, userID, carListingID int64) error {
	query := `DELETE FROM favorites WHERE user_id = ? AND car_listing_id = ?`
	result, err := r.db.ExecContext(ctx, query, userID, carListingID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrFavoriteNotFound
	}

	return nil
}

func (r *favoriteRepository) Exists(ctx context.Context, userID, carListingID int64) (bool, error) {
	query := `SELECT COUNT(*) FROM favorites WHERE user_id = ? AND car_listing_id = ?`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, carListingID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *favoriteRepository) List(ctx context.Context, filter *domain.FavoriteFilter) ([]domain.Favorite, int64, error) {
	// Count total
	countQuery := `SELECT COUNT(*) FROM favorites WHERE user_id = ?`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, filter.UserID).Scan(&total); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []domain.Favorite{}, 0, nil
	}

	// Get data
	offset := (filter.Page - 1) * filter.Limit
	query := `
		SELECT id, user_id, car_listing_id, created_at
		FROM favorites
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, filter.UserID, filter.Limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var favorites []domain.Favorite
	for rows.Next() {
		var fav domain.Favorite
		err := rows.Scan(&fav.ID, &fav.UserID, &fav.CarListingID, &fav.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		favorites = append(favorites, fav)
	}

	return favorites, total, nil
}

func (r *favoriteRepository) GetByID(ctx context.Context, id int64) (*domain.Favorite, error) {
	query := `SELECT id, user_id, car_listing_id, created_at FROM favorites WHERE id = ?`

	favorite := &domain.Favorite{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&favorite.ID, &favorite.UserID, &favorite.CarListingID, &favorite.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrFavoriteNotFound
		}
		return nil, err
	}

	return favorite, nil
}
