package sqlite

import (
	"car-marketplace/internal/domain"
	"car-marketplace/internal/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// carRepository implements repository.CarRepository
type carRepository struct {
	db *sql.DB
}

// NewCarRepository creates a new car repository instance
func NewCarRepository(db *sql.DB) repository.CarRepository {
	return &carRepository{db: db}
}

// ================= CAR LISTING CRUD =================

const listingColumns = `
	id, seller_id, make_id, model_id, year, body_type_id, fuel_type_id,
	transmission_type_id, color_id, condition_id, COALESCE(vin, ''), mileage, engine_size,
	COALESCE(horsepower, 0), COALESCE(number_of_doors, 0), COALESCE(seating_capacity, 0),
	price, currency, is_negotiable,
	title, description, city, state, country, COALESCE(postal_code, ''), status, is_featured,
	views_count, created_at, updated_at, expires_at, sold_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanListing(row rowScanner, listing *domain.CarListing) error {
	return row.Scan(
		&listing.ID, &listing.SellerID, &listing.MakeID, &listing.ModelID, &listing.Year,
		&listing.BodyTypeID, &listing.FuelTypeID, &listing.TransmissionID, &listing.ColorID,
		&listing.ConditionID, &listing.VIN, &listing.Mileage, &listing.EngineSize,
		&listing.Horsepower, &listing.NumberOfDoors, &listing.SeatingCapacity, &listing.Price,
		&listing.Currency, &listing.IsNegotiable, &listing.Title, &listing.Description,
		&listing.City, &listing.State, &listing.Country, &listing.PostalCode, &listing.Status,
		&listing.IsFeatured, &listing.ViewsCount, &listing.CreatedAt, &listing.UpdatedAt,
		&listing.ExpiresAt, &listing.SoldAt,
	)
}

// mapListingWriteError converts constraint failures into domain errors.
func mapListingWriteError(err error) error {
	switch {
	case isUniqueViolation(err, "car_listings.vin"):
		return domain.ErrDuplicateVIN
	case isForeignKeyViolation(err):
		return domain.ErrInvalidInput
	}
	return err
}

// Create creates a new car listing
func (r *carRepository) Create(ctx context.Context, listing *domain.CarListing) error {
	query := `
		INSERT INTO car_listings (
			seller_id, make_id, model_id, year, body_type_id, fuel_type_id,
			transmission_type_id, color_id, condition_id, vin, mileage, engine_size,
			horsepower, number_of_doors, seating_capacity, price, currency, is_negotiable,
			title, description, city, state, country, postal_code, status, is_featured,
			views_count, created_at, updated_at, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		listing.SellerID, listing.MakeID, listing.ModelID, listing.Year, listing.BodyTypeID,
		listing.FuelTypeID, listing.TransmissionID, listing.ColorID, listing.ConditionID,
		nullString(listing.VIN), listing.Mileage, listing.EngineSize, listing.Horsepower,
		listing.NumberOfDoors, listing.SeatingCapacity, listing.Price, listing.Currency,
		listing.IsNegotiable, listing.Title, listing.Description, listing.City, listing.State,
		listing.Country, listing.PostalCode, listing.Status, listing.IsFeatured,
		listing.ViewsCount, listing.CreatedAt, listing.UpdatedAt, listing.ExpiresAt,
	)

	if err != nil {
		if mapped := mapListingWriteError(err); mapped != err {
			return mapped
		}
		return fmt.Errorf("failed to create car listing: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	listing.ID = id
	return nil
}

// GetByID retrieves a car listing by ID
func (r *carRepository) GetByID(ctx context.Context, id int64) (*domain.CarListing, error) {
	query := `SELECT ` + listingColumns + ` FROM car_listings WHERE id = ?`

	listing := &domain.CarListing{}
	if err := scanListing(conn(ctx, r.db).QueryRowContext(ctx, query, id), listing); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCarListingNotFound
		}
		return nil, fmt.Errorf("failed to get car listing: %w", err)
	}

	return listing, nil
}

// Update updates an existing car listing
func (r *carRepository) Update(ctx context.Context, listing *domain.CarListing) error {
	query := `
		UPDATE car_listings SET
			make_id = ?, model_id = ?, year = ?, body_type_id = ?, fuel_type_id = ?,
			transmission_type_id = ?, color_id = ?, condition_id = ?, vin = ?, mileage = ?,
			engine_size = ?, horsepower = ?, number_of_doors = ?, seating_capacity = ?,
			price = ?, currency = ?, is_negotiable = ?, title = ?, description = ?,
			city = ?, state = ?, country = ?, postal_code = ?, status = ?, is_featured = ?,
			updated_at = ?, expires_at = ?, sold_at = ?
		WHERE id = ?
	`

	result, err := conn(ctx, r.db).ExecContext(ctx, query,
		listing.MakeID, listing.ModelID, listing.Year, listing.BodyTypeID, listing.FuelTypeID,
		listing.TransmissionID, listing.ColorID, listing.ConditionID, nullString(listing.VIN), listing.Mileage,
		listing.EngineSize, listing.Horsepower, listing.NumberOfDoors, listing.SeatingCapacity,
		listing.Price, listing.Currency, listing.IsNegotiable, listing.Title, listing.Description,
		listing.City, listing.State, listing.Country, listing.PostalCode, listing.Status,
		listing.IsFeatured, listing.UpdatedAt, listing.ExpiresAt, listing.SoldAt, listing.ID,
	)

	if err != nil {
		if mapped := mapListingWriteError(err); mapped != err {
			return mapped
		}
		return fmt.Errorf("failed to update car listing: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrCarListingNotFound
	}

	return nil
}

// Delete deletes a car listing by ID
func (r *carRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM car_listings WHERE id = ?`

	result, err := conn(ctx, r.db).ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete car listing: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrCarListingNotFound
	}

	return nil
}

// ftsQuery turns free text into a safe FTS5 query: each word becomes a quoted
// prefix term ("word"*), and all terms must match. Punctuation and FTS
// operators in the input are dropped, so user input can't break the syntax.
func ftsQuery(search string) string {
	words := strings.FieldsFunc(search, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	terms := make([]string, 0, len(words))
	for _, w := range words {
		terms = append(terms, `"`+w+`"*`)
	}
	return strings.Join(terms, " ")
}

// List retrieves car listings based on filter criteria
func (r *carRepository) List(ctx context.Context, filter *domain.CarListingFilter) ([]domain.CarListing, int64, error) {
	if filter == nil {
		filter = &domain.CarListingFilter{}
	}
	filter.SetDefaults()

	// Build WHERE clause
	where := []string{"1=1"}
	args := []interface{}{}

	add := func(cond string, arg any) {
		where = append(where, cond)
		args = append(args, arg)
	}

	if filter.MakeID != nil {
		add("make_id = ?", *filter.MakeID)
	}
	if filter.ModelID != nil {
		add("model_id = ?", *filter.ModelID)
	}
	if filter.MinYear != nil {
		add("year >= ?", *filter.MinYear)
	}
	if filter.MaxYear != nil {
		add("year <= ?", *filter.MaxYear)
	}
	if filter.BodyTypeID != nil {
		add("body_type_id = ?", *filter.BodyTypeID)
	}
	if filter.FuelTypeID != nil {
		add("fuel_type_id = ?", *filter.FuelTypeID)
	}
	if filter.TransmissionID != nil {
		add("transmission_type_id = ?", *filter.TransmissionID)
	}
	if filter.ColorID != nil {
		add("color_id = ?", *filter.ColorID)
	}
	if filter.ConditionID != nil {
		add("condition_id = ?", *filter.ConditionID)
	}
	if filter.MinPrice != nil {
		add("price >= ?", *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		add("price <= ?", *filter.MaxPrice)
	}
	if filter.MinMileage != nil {
		add("mileage >= ?", *filter.MinMileage)
	}
	if filter.MaxMileage != nil {
		add("mileage <= ?", *filter.MaxMileage)
	}
	// COLLATE NOCASE matches idx_car_listings_location, so these use the index.
	if filter.City != nil {
		add("city = ? COLLATE NOCASE", *filter.City)
	}
	if filter.State != nil {
		add("state = ? COLLATE NOCASE", *filter.State)
	}
	if filter.Country != nil {
		add("country = ? COLLATE NOCASE", *filter.Country)
	}
	if filter.Status != nil {
		add("status = ?", *filter.Status)
	}
	if filter.SellerID != nil {
		add("seller_id = ?", *filter.SellerID)
	}
	if filter.IsFeatured != nil {
		add("is_featured = ?", *filter.IsFeatured)
	}
	if filter.Search != nil {
		if q := ftsQuery(*filter.Search); q != "" {
			add("id IN (SELECT rowid FROM car_listings_fts WHERE car_listings_fts MATCH ?)", q)
		}
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM car_listings WHERE %s", whereClause)
	var total int64
	err := conn(ctx, r.db).QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count car listings: %w", err)
	}

	listings := []domain.CarListing{}
	if total == 0 {
		return listings, 0, nil
	}

	// Build ORDER BY clause; id breaks ties so pages don't overlap.
	orderBy := "created_at DESC, id DESC"
	switch filter.SortBy {
	case "price_asc":
		orderBy = "price ASC, id ASC"
	case "price_desc":
		orderBy = "price DESC, id DESC"
	case "year_asc":
		orderBy = "year ASC, id ASC"
	case "year_desc":
		orderBy = "year DESC, id DESC"
	case "mileage_asc":
		orderBy = "mileage ASC, id ASC"
	case "mileage_desc":
		orderBy = "mileage DESC, id DESC"
	case "created_asc":
		orderBy = "created_at ASC, id ASC"
	}

	// Build pagination
	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	query := fmt.Sprintf(`SELECT %s FROM car_listings WHERE %s ORDER BY %s LIMIT ? OFFSET ?`,
		listingColumns, whereClause, orderBy)

	rows, err := conn(ctx, r.db).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list car listings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var listing domain.CarListing
		if err := scanListing(rows, &listing); err != nil {
			return nil, 0, fmt.Errorf("failed to scan car listing: %w", err)
		}
		listings = append(listings, listing)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating car listings: %w", err)
	}

	return listings, total, nil
}

// IncrementViews adds buffered view counts to their listings in one transaction.
func (r *carRepository) IncrementViews(ctx context.Context, counts map[int64]int) error {
	if len(counts) == 0 {
		return nil
	}

	return withTx(ctx, r.db, func(ctx context.Context, q querier) error {
		for id, n := range counts {
			if _, err := q.ExecContext(ctx, `UPDATE car_listings SET views_count = views_count + ? WHERE id = ?`, n, id); err != nil {
				return fmt.Errorf("failed to increment views: %w", err)
			}
		}
		return nil
	})
}

// ================= REFERENCE DATA =================

// queryAll runs query and scans each row with scan, collecting the results.
func queryAll[T any](ctx context.Context, q querier, query string, scan func(rowScanner, *T) error, args ...any) ([]T, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []T{}
	for rows.Next() {
		var item T
		if err := scan(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *carRepository) GetAllMakes(ctx context.Context) ([]domain.CarMake, error) {
	return queryAll(ctx, conn(ctx, r.db),
		`SELECT id, name, COALESCE(logo_url, ''), COALESCE(country_of_origin, ''), created_at, updated_at FROM car_makes ORDER BY name ASC`,
		func(row rowScanner, m *domain.CarMake) error {
			return row.Scan(&m.ID, &m.Name, &m.LogoURL, &m.CountryOfOrigin, &m.CreatedAt, &m.UpdatedAt)
		})
}

func (r *carRepository) GetAllModels(ctx context.Context) ([]domain.CarModel, error) {
	return queryAll(ctx, conn(ctx, r.db),
		`SELECT id, make_id, name, created_at, updated_at FROM car_models ORDER BY name ASC`,
		func(row rowScanner, m *domain.CarModel) error {
			return row.Scan(&m.ID, &m.MakeID, &m.Name, &m.CreatedAt, &m.UpdatedAt)
		})
}

func (r *carRepository) GetAllBodyTypes(ctx context.Context) ([]domain.BodyType, error) {
	return queryAll(ctx, conn(ctx, r.db), `SELECT id, name, created_at FROM body_types ORDER BY name ASC`,
		func(row rowScanner, b *domain.BodyType) error { return row.Scan(&b.ID, &b.Name, &b.CreatedAt) })
}

func (r *carRepository) GetAllFuelTypes(ctx context.Context) ([]domain.FuelType, error) {
	return queryAll(ctx, conn(ctx, r.db), `SELECT id, name, created_at FROM fuel_types ORDER BY name ASC`,
		func(row rowScanner, f *domain.FuelType) error { return row.Scan(&f.ID, &f.Name, &f.CreatedAt) })
}

func (r *carRepository) GetAllTransmissions(ctx context.Context) ([]domain.TransmissionType, error) {
	return queryAll(ctx, conn(ctx, r.db), `SELECT id, name, created_at FROM transmission_types ORDER BY name ASC`,
		func(row rowScanner, t *domain.TransmissionType) error { return row.Scan(&t.ID, &t.Name, &t.CreatedAt) })
}

func (r *carRepository) GetAllColors(ctx context.Context) ([]domain.Color, error) {
	return queryAll(ctx, conn(ctx, r.db), `SELECT id, name, COALESCE(hex_code, ''), created_at FROM colors ORDER BY name ASC`,
		func(row rowScanner, c *domain.Color) error { return row.Scan(&c.ID, &c.Name, &c.HexCode, &c.CreatedAt) })
}

func (r *carRepository) GetAllConditions(ctx context.Context) ([]domain.CarCondition, error) {
	return queryAll(ctx, conn(ctx, r.db), `SELECT id, name, created_at FROM car_conditions ORDER BY name ASC`,
		func(row rowScanner, c *domain.CarCondition) error { return row.Scan(&c.ID, &c.Name, &c.CreatedAt) })
}

func scanFeature(row rowScanner, f *domain.Feature) error {
	return row.Scan(&f.ID, &f.Name, &f.Category, &f.CreatedAt)
}

func (r *carRepository) GetAllFeatures(ctx context.Context) ([]domain.Feature, error) {
	return queryAll(ctx, conn(ctx, r.db),
		`SELECT id, name, COALESCE(category, ''), created_at FROM features ORDER BY category, name ASC`, scanFeature)
}

// ================= FEATURE OPERATIONS =================

// ReplaceListingFeatures sets a listing's features to exactly featureIDs.
func (r *carRepository) ReplaceListingFeatures(ctx context.Context, listingID int64, featureIDs []int64) error {
	return withTx(ctx, r.db, func(ctx context.Context, q querier) error {
		if _, err := q.ExecContext(ctx, `DELETE FROM car_listing_features WHERE car_listing_id = ?`, listingID); err != nil {
			return err
		}

		for _, featureID := range featureIDs {
			_, err := q.ExecContext(ctx,
				`INSERT OR IGNORE INTO car_listing_features (car_listing_id, feature_id, created_at) VALUES (?, ?, datetime('now'))`,
				listingID, featureID)
			if err != nil {
				if isForeignKeyViolation(err) {
					return domain.ErrInvalidFeature
				}
				return err
			}
		}
		return nil
	})
}

func (r *carRepository) GetListingFeatures(ctx context.Context, listingID int64) ([]domain.Feature, error) {
	return queryAll(ctx, conn(ctx, r.db), `
		SELECT f.id, f.name, COALESCE(f.category, ''), f.created_at
		FROM features f
		INNER JOIN car_listing_features clf ON f.id = clf.feature_id
		WHERE clf.car_listing_id = ?
		ORDER BY f.category, f.name ASC
	`, scanFeature, listingID)
}

// ================= IMAGE OPERATIONS =================

const imageColumns = `id, car_listing_id, image_url, display_order, is_primary, created_at`

func scanImage(row rowScanner, img *domain.CarImage) error {
	return row.Scan(&img.ID, &img.CarListingID, &img.ImageURL, &img.DisplayOrder, &img.IsPrimary, &img.CreatedAt)
}

// CreateImage appends an image to its listing. The display order is set after
// the listing's last image, and the first image becomes the primary one.
func (r *carRepository) CreateImage(ctx context.Context, image *domain.CarImage) error {
	return withTx(ctx, r.db, func(ctx context.Context, q querier) error {
		var count int
		err := q.QueryRowContext(ctx,
			`SELECT COALESCE(MAX(display_order) + 1, 0), COUNT(*) FROM car_images WHERE car_listing_id = ?`,
			image.CarListingID,
		).Scan(&image.DisplayOrder, &count)
		if err != nil {
			return err
		}
		image.IsPrimary = count == 0

		result, err := q.ExecContext(ctx, `
			INSERT INTO car_images (car_listing_id, image_url, display_order, is_primary, created_at)
			VALUES (?, ?, ?, ?, ?)`,
			image.CarListingID, image.ImageURL, image.DisplayOrder, image.IsPrimary, image.CreatedAt,
		)
		if err != nil {
			if isForeignKeyViolation(err) {
				return domain.ErrCarListingNotFound
			}
			return err
		}

		image.ID, err = result.LastInsertId()
		return err
	})
}

func (r *carRepository) GetImageByID(ctx context.Context, id int64) (*domain.CarImage, error) {
	img := &domain.CarImage{}
	err := scanImage(conn(ctx, r.db).QueryRowContext(ctx, `SELECT `+imageColumns+` FROM car_images WHERE id = ?`, id), img)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrImageNotFound
		}
		return nil, err
	}
	return img, nil
}

func (r *carRepository) GetListingImages(ctx context.Context, listingID int64) ([]domain.CarImage, error) {
	return queryAll(ctx, conn(ctx, r.db), `
		SELECT `+imageColumns+`
		FROM car_images
		WHERE car_listing_id = ?
		ORDER BY is_primary DESC, display_order ASC
	`, scanImage, listingID)
}

// DeleteImage removes an image; if it was the primary image, the next image
// in display order becomes primary.
func (r *carRepository) DeleteImage(ctx context.Context, id int64) error {
	return withTx(ctx, r.db, func(ctx context.Context, q querier) error {
		var listingID int64
		var wasPrimary bool
		err := q.QueryRowContext(ctx, `DELETE FROM car_images WHERE id = ? RETURNING car_listing_id, is_primary`, id).
			Scan(&listingID, &wasPrimary)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrImageNotFound
			}
			return err
		}

		if wasPrimary {
			_, err = q.ExecContext(ctx, `
				UPDATE car_images SET is_primary = 1
				WHERE id = (SELECT id FROM car_images WHERE car_listing_id = ? ORDER BY display_order, id LIMIT 1)`,
				listingID)
		}
		return err
	})
}

// SetPrimaryImage makes imageID the listing's only primary image. It fails
// with domain.ErrImageNotFound if the image doesn't belong to the listing.
func (r *carRepository) SetPrimaryImage(ctx context.Context, listingID int64, imageID int64) error {
	return withTx(ctx, r.db, func(ctx context.Context, q querier) error {
		var exists int
		err := q.QueryRowContext(ctx, `SELECT COUNT(*) FROM car_images WHERE id = ? AND car_listing_id = ?`, imageID, listingID).Scan(&exists)
		if err != nil {
			return err
		}
		if exists == 0 {
			return domain.ErrImageNotFound
		}

		_, err = q.ExecContext(ctx, `UPDATE car_images SET is_primary = (id = ?) WHERE car_listing_id = ?`, imageID, listingID)
		return err
	})
}
