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

// carRepository implements repository.CarRepository
type carRepository struct {
	db *sql.DB
}

// NewCarRepository creates a new car repository instance
func NewCarRepository(db *sql.DB) repository.CarRepository {
	return &carRepository{db: db}
}

// ================= CAR LISTING CRUD =================

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

	result, err := r.db.ExecContext(ctx, query,
		listing.SellerID, listing.MakeID, listing.ModelID, listing.Year, listing.BodyTypeID,
		listing.FuelTypeID, listing.TransmissionID, listing.ColorID, listing.ConditionID,
		listing.VIN, listing.Mileage, listing.EngineSize, listing.Horsepower,
		listing.NumberOfDoors, listing.SeatingCapacity, listing.Price, listing.Currency,
		listing.IsNegotiable, listing.Title, listing.Description, listing.City, listing.State,
		listing.Country, listing.PostalCode, listing.Status, listing.IsFeatured,
		listing.ViewsCount, listing.CreatedAt, listing.UpdatedAt, listing.ExpiresAt,
	)

	if err != nil {
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
	query := `
		SELECT
			id, seller_id, make_id, model_id, year, body_type_id, fuel_type_id,
			transmission_type_id, color_id, condition_id, COALESCE(vin, ''), mileage, engine_size,
			COALESCE(horsepower, 0), COALESCE(number_of_doors, 0), COALESCE(seating_capacity, 0),
			price, currency, is_negotiable,
			title, description, city, state, country, COALESCE(postal_code, ''), status, is_featured,
			views_count, created_at, updated_at, expires_at, sold_at
		FROM car_listings
		WHERE id = ?
	`

	listing := &domain.CarListing{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&listing.ID, &listing.SellerID, &listing.MakeID, &listing.ModelID, &listing.Year,
		&listing.BodyTypeID, &listing.FuelTypeID, &listing.TransmissionID, &listing.ColorID,
		&listing.ConditionID, &listing.VIN, &listing.Mileage, &listing.EngineSize,
		&listing.Horsepower, &listing.NumberOfDoors, &listing.SeatingCapacity, &listing.Price,
		&listing.Currency, &listing.IsNegotiable, &listing.Title, &listing.Description,
		&listing.City, &listing.State, &listing.Country, &listing.PostalCode, &listing.Status,
		&listing.IsFeatured, &listing.ViewsCount, &listing.CreatedAt, &listing.UpdatedAt,
		&listing.ExpiresAt, &listing.SoldAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
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

	result, err := r.db.ExecContext(ctx, query,
		listing.MakeID, listing.ModelID, listing.Year, listing.BodyTypeID, listing.FuelTypeID,
		listing.TransmissionID, listing.ColorID, listing.ConditionID, listing.VIN, listing.Mileage,
		listing.EngineSize, listing.Horsepower, listing.NumberOfDoors, listing.SeatingCapacity,
		listing.Price, listing.Currency, listing.IsNegotiable, listing.Title, listing.Description,
		listing.City, listing.State, listing.Country, listing.PostalCode, listing.Status,
		listing.IsFeatured, listing.UpdatedAt, listing.ExpiresAt, listing.SoldAt, listing.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update car listing: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete deletes a car listing by ID
func (r *carRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM car_listings WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete car listing: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
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

	if filter.MakeID != nil {
		where = append(where, "make_id = ?")
		args = append(args, *filter.MakeID)
	}
	if filter.ModelID != nil {
		where = append(where, "model_id = ?")
		args = append(args, *filter.ModelID)
	}
	if filter.MinYear != nil {
		where = append(where, "year >= ?")
		args = append(args, *filter.MinYear)
	}
	if filter.MaxYear != nil {
		where = append(where, "year <= ?")
		args = append(args, *filter.MaxYear)
	}
	if filter.BodyTypeID != nil {
		where = append(where, "body_type_id = ?")
		args = append(args, *filter.BodyTypeID)
	}
	if filter.FuelTypeID != nil {
		where = append(where, "fuel_type_id = ?")
		args = append(args, *filter.FuelTypeID)
	}
	if filter.TransmissionID != nil {
		where = append(where, "transmission_type_id = ?")
		args = append(args, *filter.TransmissionID)
	}
	if filter.ColorID != nil {
		where = append(where, "color_id = ?")
		args = append(args, *filter.ColorID)
	}
	if filter.ConditionID != nil {
		where = append(where, "condition_id = ?")
		args = append(args, *filter.ConditionID)
	}
	if filter.MinPrice != nil {
		where = append(where, "price >= ?")
		args = append(args, *filter.MinPrice)
	}
	if filter.MaxPrice != nil {
		where = append(where, "price <= ?")
		args = append(args, *filter.MaxPrice)
	}
	if filter.MinMileage != nil {
		where = append(where, "mileage >= ?")
		args = append(args, *filter.MinMileage)
	}
	if filter.MaxMileage != nil {
		where = append(where, "mileage <= ?")
		args = append(args, *filter.MaxMileage)
	}
	if filter.City != nil {
		where = append(where, "LOWER(city) = LOWER(?)")
		args = append(args, *filter.City)
	}
	if filter.State != nil {
		where = append(where, "LOWER(state) = LOWER(?)")
		args = append(args, *filter.State)
	}
	if filter.Country != nil {
		where = append(where, "LOWER(country) = LOWER(?)")
		args = append(args, *filter.Country)
	}
	if filter.Status != nil {
		where = append(where, "status = ?")
		args = append(args, *filter.Status)
	}
	if filter.SellerID != nil {
		where = append(where, "seller_id = ?")
		args = append(args, *filter.SellerID)
	}
	if filter.IsFeatured != nil {
		where = append(where, "is_featured = ?")
		args = append(args, *filter.IsFeatured)
	}
	if filter.Search != nil && *filter.Search != "" {
		where = append(where, "(LOWER(title) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?))")
		searchTerm := "%" + *filter.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM car_listings WHERE %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count car listings: %w", err)
	}

	// Build ORDER BY clause
	orderBy := "created_at DESC"
	switch filter.SortBy {
	case "price_asc":
		orderBy = "price ASC"
	case "price_desc":
		orderBy = "price DESC"
	case "year_asc":
		orderBy = "year ASC"
	case "year_desc":
		orderBy = "year DESC"
	case "mileage_asc":
		orderBy = "mileage ASC"
	case "mileage_desc":
		orderBy = "mileage DESC"
	case "created_asc":
		orderBy = "created_at ASC"
	case "created_desc":
		orderBy = "created_at DESC"
	}

	// Build pagination
	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	// Get listings
	query := fmt.Sprintf(`
		SELECT
			id, seller_id, make_id, model_id, year, body_type_id, fuel_type_id,
			transmission_type_id, color_id, condition_id, COALESCE(vin, ''), mileage, engine_size,
			COALESCE(horsepower, 0), COALESCE(number_of_doors, 0), COALESCE(seating_capacity, 0),
			price, currency, is_negotiable,
			title, description, city, state, country, COALESCE(postal_code, ''), status, is_featured,
			views_count, created_at, updated_at, expires_at, sold_at
		FROM car_listings
		WHERE %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, whereClause, orderBy)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list car listings: %w", err)
	}
	defer rows.Close()

	listings := []domain.CarListing{}
	for rows.Next() {
		var listing domain.CarListing
		err := rows.Scan(
			&listing.ID, &listing.SellerID, &listing.MakeID, &listing.ModelID, &listing.Year,
			&listing.BodyTypeID, &listing.FuelTypeID, &listing.TransmissionID, &listing.ColorID,
			&listing.ConditionID, &listing.VIN, &listing.Mileage, &listing.EngineSize,
			&listing.Horsepower, &listing.NumberOfDoors, &listing.SeatingCapacity, &listing.Price,
			&listing.Currency, &listing.IsNegotiable, &listing.Title, &listing.Description,
			&listing.City, &listing.State, &listing.Country, &listing.PostalCode, &listing.Status,
			&listing.IsFeatured, &listing.ViewsCount, &listing.CreatedAt, &listing.UpdatedAt,
			&listing.ExpiresAt, &listing.SoldAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan car listing: %w", err)
		}
		listings = append(listings, listing)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating car listings: %w", err)
	}

	return listings, total, nil
}

// IncrementViews increments the view count for a car listing
func (r *carRepository) IncrementViews(ctx context.Context, id int64) error {
	query := `UPDATE car_listings SET views_count = views_count + 1 WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment views: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ================= CAR MAKE OPERATIONS =================

func (r *carRepository) GetMakeByID(ctx context.Context, id int64) (*domain.CarMake, error) {
	query := `SELECT id, name, COALESCE(logo_url, ''), COALESCE(country_of_origin, ''), created_at, updated_at FROM car_makes WHERE id = ?`

	make := &domain.CarMake{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&make.ID, &make.Name, &make.LogoURL, &make.CountryOfOrigin, &make.CreatedAt, &make.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidMake
		}
		return nil, err
	}

	return make, nil
}

func (r *carRepository) GetAllMakes(ctx context.Context) ([]domain.CarMake, error) {
	query := `SELECT id, name, COALESCE(logo_url, ''), COALESCE(country_of_origin, ''), created_at, updated_at FROM car_makes ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var makes []domain.CarMake
	for rows.Next() {
		var make domain.CarMake
		err := rows.Scan(&make.ID, &make.Name, &make.LogoURL, &make.CountryOfOrigin, &make.CreatedAt, &make.UpdatedAt)
		if err != nil {
			return nil, err
		}
		makes = append(makes, make)
	}

	return makes, nil
}

// ================= CAR MODEL OPERATIONS =================

func (r *carRepository) GetModelByID(ctx context.Context, id int64) (*domain.CarModel, error) {
	query := `SELECT id, make_id, name, created_at, updated_at FROM car_models WHERE id = ?`

	model := &domain.CarModel{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&model.ID, &model.MakeID, &model.Name, &model.CreatedAt, &model.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidModel
		}
		return nil, err
	}

	return model, nil
}

func (r *carRepository) GetModelsByMakeID(ctx context.Context, makeID int64) ([]domain.CarModel, error) {
	query := `SELECT id, make_id, name, created_at, updated_at FROM car_models WHERE make_id = ? ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, makeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []domain.CarModel
	for rows.Next() {
		var model domain.CarModel
		err := rows.Scan(&model.ID, &model.MakeID, &model.Name, &model.CreatedAt, &model.UpdatedAt)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

// ================= BODY TYPE OPERATIONS =================

func (r *carRepository) GetBodyTypeByID(ctx context.Context, id int64) (*domain.BodyType, error) {
	query := `SELECT id, name, created_at FROM body_types WHERE id = ?`

	bodyType := &domain.BodyType{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&bodyType.ID, &bodyType.Name, &bodyType.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidBodyType
		}
		return nil, err
	}

	return bodyType, nil
}

func (r *carRepository) GetAllBodyTypes(ctx context.Context) ([]domain.BodyType, error) {
	query := `SELECT id, name, created_at FROM body_types ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bodyTypes []domain.BodyType
	for rows.Next() {
		var bt domain.BodyType
		err := rows.Scan(&bt.ID, &bt.Name, &bt.CreatedAt)
		if err != nil {
			return nil, err
		}
		bodyTypes = append(bodyTypes, bt)
	}

	return bodyTypes, nil
}

// ================= FUEL TYPE OPERATIONS =================

func (r *carRepository) GetFuelTypeByID(ctx context.Context, id int64) (*domain.FuelType, error) {
	query := `SELECT id, name, created_at FROM fuel_types WHERE id = ?`

	fuelType := &domain.FuelType{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&fuelType.ID, &fuelType.Name, &fuelType.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidFuelType
		}
		return nil, err
	}

	return fuelType, nil
}

func (r *carRepository) GetAllFuelTypes(ctx context.Context) ([]domain.FuelType, error) {
	query := `SELECT id, name, created_at FROM fuel_types ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fuelTypes []domain.FuelType
	for rows.Next() {
		var ft domain.FuelType
		err := rows.Scan(&ft.ID, &ft.Name, &ft.CreatedAt)
		if err != nil {
			return nil, err
		}
		fuelTypes = append(fuelTypes, ft)
	}

	return fuelTypes, nil
}

// ================= TRANSMISSION OPERATIONS =================

func (r *carRepository) GetTransmissionByID(ctx context.Context, id int64) (*domain.TransmissionType, error) {
	query := `SELECT id, name, created_at FROM transmission_types WHERE id = ?`

	transmission := &domain.TransmissionType{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&transmission.ID, &transmission.Name, &transmission.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidTransmission
		}
		return nil, err
	}

	return transmission, nil
}

func (r *carRepository) GetAllTransmissions(ctx context.Context) ([]domain.TransmissionType, error) {
	query := `SELECT id, name, created_at FROM transmission_types ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transmissions []domain.TransmissionType
	for rows.Next() {
		var t domain.TransmissionType
		err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt)
		if err != nil {
			return nil, err
		}
		transmissions = append(transmissions, t)
	}

	return transmissions, nil
}

// ================= COLOR OPERATIONS =================

func (r *carRepository) GetColorByID(ctx context.Context, id int64) (*domain.Color, error) {
	query := `SELECT id, name, COALESCE(hex_code, ''), created_at FROM colors WHERE id = ?`

	color := &domain.Color{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&color.ID, &color.Name, &color.HexCode, &color.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidColor
		}
		return nil, err
	}

	return color, nil
}

func (r *carRepository) GetAllColors(ctx context.Context) ([]domain.Color, error) {
	query := `SELECT id, name, COALESCE(hex_code, ''), created_at FROM colors ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var colors []domain.Color
	for rows.Next() {
		var c domain.Color
		err := rows.Scan(&c.ID, &c.Name, &c.HexCode, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		colors = append(colors, c)
	}

	return colors, nil
}

// ================= CONDITION OPERATIONS =================

func (r *carRepository) GetConditionByID(ctx context.Context, id int64) (*domain.CarCondition, error) {
	query := `SELECT id, name, created_at FROM car_conditions WHERE id = ?`

	condition := &domain.CarCondition{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&condition.ID, &condition.Name, &condition.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidCondition
		}
		return nil, err
	}

	return condition, nil
}

func (r *carRepository) GetAllConditions(ctx context.Context) ([]domain.CarCondition, error) {
	query := `SELECT id, name, created_at FROM car_conditions ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conditions []domain.CarCondition
	for rows.Next() {
		var c domain.CarCondition
		err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, c)
	}

	return conditions, nil
}

// ================= FEATURE OPERATIONS =================

func (r *carRepository) GetFeatureByID(ctx context.Context, id int64) (*domain.Feature, error) {
	query := `SELECT id, name, COALESCE(category, ''), created_at FROM features WHERE id = ?`

	feature := &domain.Feature{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&feature.ID, &feature.Name, &feature.Category, &feature.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidFeature
		}
		return nil, err
	}

	return feature, nil
}

func (r *carRepository) GetAllFeatures(ctx context.Context) ([]domain.Feature, error) {
	query := `SELECT id, name, COALESCE(category, ''), created_at FROM features ORDER BY category, name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []domain.Feature
	for rows.Next() {
		var f domain.Feature
		err := rows.Scan(&f.ID, &f.Name, &f.Category, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		features = append(features, f)
	}

	return features, nil
}

func (r *carRepository) GetFeaturesByCategory(ctx context.Context, category string) ([]domain.Feature, error) {
	query := `SELECT id, name, COALESCE(category, ''), created_at FROM features WHERE category = ? ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []domain.Feature
	for rows.Next() {
		var f domain.Feature
		err := rows.Scan(&f.ID, &f.Name, &f.Category, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		features = append(features, f)
	}

	return features, nil
}

func (r *carRepository) AddFeaturesToListing(ctx context.Context, listingID int64, featureIDs []int64) error {
	if len(featureIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO car_listing_features (car_listing_id, feature_id, created_at) VALUES (?, ?, datetime('now'))`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, featureID := range featureIDs {
		if _, err := stmt.ExecContext(ctx, listingID, featureID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *carRepository) RemoveAllFeaturesFromListing(ctx context.Context, listingID int64) error {
	query := `DELETE FROM car_listing_features WHERE car_listing_id = ?`
	_, err := r.db.ExecContext(ctx, query, listingID)
	return err
}

func (r *carRepository) GetListingFeatures(ctx context.Context, listingID int64) ([]domain.Feature, error) {
	query := `
		SELECT f.id, f.name, COALESCE(f.category, ''), f.created_at
		FROM features f
		INNER JOIN car_listing_features clf ON f.id = clf.feature_id
		WHERE clf.car_listing_id = ?
		ORDER BY f.category, f.name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []domain.Feature
	for rows.Next() {
		var f domain.Feature
		err := rows.Scan(&f.ID, &f.Name, &f.Category, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		features = append(features, f)
	}

	return features, nil
}

// ================= IMAGE OPERATIONS =================

func (r *carRepository) CreateImage(ctx context.Context, image *domain.CarImage) error {
	query := `
		INSERT INTO car_images (car_listing_id, image_url, display_order, is_primary, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		image.CarListingID, image.ImageURL, image.DisplayOrder, image.IsPrimary, image.CreatedAt,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	image.ID = id
	return nil
}

func (r *carRepository) GetListingImages(ctx context.Context, listingID int64) ([]domain.CarImage, error) {
	query := `
		SELECT id, car_listing_id, image_url, display_order, is_primary, created_at
		FROM car_images
		WHERE car_listing_id = ?
		ORDER BY is_primary DESC, display_order ASC
	`

	rows, err := r.db.QueryContext(ctx, query, listingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []domain.CarImage
	for rows.Next() {
		var img domain.CarImage
		err := rows.Scan(&img.ID, &img.CarListingID, &img.ImageURL, &img.DisplayOrder, &img.IsPrimary, &img.CreatedAt)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}

	return images, nil
}

func (r *carRepository) DeleteImage(ctx context.Context, id int64) error {
	query := `DELETE FROM car_images WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *carRepository) SetPrimaryImage(ctx context.Context, listingID int64, imageID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear all primary flags for this listing
	if _, err := tx.ExecContext(ctx, `UPDATE car_images SET is_primary = 0 WHERE car_listing_id = ?`, listingID); err != nil {
		return err
	}

	// Set new primary image
	if _, err := tx.ExecContext(ctx, `UPDATE car_images SET is_primary = 1 WHERE id = ? AND car_listing_id = ?`, imageID, listingID); err != nil {
		return err
	}

	return tx.Commit()
}
