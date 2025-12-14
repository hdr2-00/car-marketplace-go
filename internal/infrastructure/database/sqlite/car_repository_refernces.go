package sqlite

// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"car-marketplace/internal/domain"
// )

// // Make operations

// // func (r *carRepository) GetMakeByID(ctx context.Context, id int64) (*domain.CarMake, error) {
// // 	query := `SELECT id, name, logo_url, country_of_origin, created_at, updated_at FROM car_makes WHERE id = ?`

// // 	make := &domain.CarMake{}
// // 	err := r.db.QueryRowContext(ctx, query, id).Scan(
// // 		&make.ID, &make.Name, &make.LogoURL, &make.CountryOfOrigin, &make.CreatedAt, &make.UpdatedAt,
// // 	)

// // 	if err != nil {
// // 		if errors.Is(err, sql.ErrNoRows) {
// // 			return features, nil
// // 		}
// // 	}
// // // TODO: Re-implement AddFeaturesToListing if needed in the future
// // func (r *carRepository) AddFeaturesToListing(ctx context.Context, listingID int64, featureIDs []int64) error {
// // 	if len(featureIDs) == 0 {
// // 		return nil
// // 	}

// // 	tx, err := r.db.BeginTx(ctx, nil)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	defer tx.Rollback()

// // 	stmt, err := tx.PrepareContext(ctx, `INSERT INTO car_listing_features (car_listing_id, feature_id, created_at) VALUES (?, ?, datetime('now'))`)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	defer stmt.Close()

// // 	for _, featureID := range featureIDs {
// // 		if _, err := stmt.ExecContext(ctx, listingID, featureID); err != nil {
// // 			return err
// // 		}
// // 	}

// // 	return tx.Commit()
// // }

// // func (r *carRepository) RemoveAllFeaturesFromListing(ctx context.Context, listingID int64) error {
// // 	query := `DELETE FROM car_listing_features WHERE car_listing_id = ?`
// // 	_, err := r.db.ExecContext(ctx, query, listingID)
// // 	return err
// // }

// // // func (r *carRepository) GetListingFeatures(ctx context.Context, listingID int64) ([]domain.Feature, error) {
// // 	query := `
// // 		SELECT f.id, f.name, f.category, f.created_at
// // 		FROM features f
// // 		INNER JOIN car_listing_features clf ON f.id = clf.feature_id
// // 		WHERE clf.car_listing_id = ?
// // 		ORDER BY f.category, f.name ASC
// // 	`

// // 	rows, err := r.db.QueryContext(ctx, query, listingID)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	defer rows.Close()

// // 	var features []domain.Feature
// // 	for rows.Next() {
// // 		var f domain.Feature
// // 		err := rows.Scan(&f.ID, &f.Name, &f.Category, &f.CreatedAt)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		features = append(features, f)
// // 	}

// // 	return features, nil
// // }

// // // Image operations

// // func (r *carRepository) CreateImage(ctx context.Context, image *domain.CarImage) error {
// // 	query := `
// // 		INSERT INTO car_images (car_listing_id, image_url, display_order, is_primary, created_at)
// // 		VALUES (?, ?, ?, ?, ?)
// // 	`

// // 	result, err := r.db.ExecContext(ctx, query,
// // 		image.CarListingID, image.ImageURL, image.DisplayOrder, image.IsPrimary, image.CreatedAt,
// // 	)

// // 	if err != nil {
// // 		return err
// // 	}

// // 	id, err := result.LastInsertId()
// // 	if err != nil {
// // 		return err
// // 	}

// // 	image.ID = id
// // 	return nil
// // }

// // func (r *carRepository) GetListingImages(ctx context.Context, listingID int64) ([]domain.CarImage, error) {
// // 	query := `
// // 		SELECT id, car_listing_id, image_url, display_order, is_primary, created_at
// // 		FROM car_images
// // 		WHERE car_listing_id = ?
// // 		ORDER BY is_primary DESC, display_order ASC
// // 	`

// // 	rows, err := r.db.QueryContext(ctx, query, listingID)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	defer rows.Close()

// // 	var images []domain.CarImage
// // 	for rows.Next() {
// // 		var img domain.CarImage
// // 		err := rows.Scan(&img.ID, &img.CarListingID, &img.ImageURL, &img.DisplayOrder, &img.IsPrimary, &img.CreatedAt)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		images = append(images, img)
// // 	}

// // 	return images, nil
// // }

// // func (r *carRepository) DeleteImage(ctx context.Context, id int64) error {
// // 	query := `DELETE FROM car_images WHERE id = ?`
// // 	result, err := r.db.ExecContext(ctx, query, id)
// // 	if err != nil {
// // 		return err
// // 	}

// // 	rowsAffected, err := result.RowsAffected()
// // 	if err != nil {
// // 		return err
// // 	}

// // 	if rowsAffected == 0 {
// // 		return domain.ErrNotFound
// // 	}

// // 	return nil
// // }

// // func (r *carRepository) SetPrimaryImage(ctx context.Context, listingID int64, imageID int64) error {
// // 	tx, err := r.db.BeginTx(ctx, nil)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	defer tx.Rollback()

// // 	// Clear all primary flags for this listing
// // 	if _, err := tx.ExecContext(ctx, `UPDATE car_images SET is_primary = 0 WHERE car_listing_id = ?`, listingID); err != nil {
// // 		return err
// // 	}

// // 	// Set new primary image
// // 	if _, err := tx.ExecContext(ctx, `UPDATE car_images SET is_primary = 1 WHERE id = ? AND car_listing_id = ?`, imageID, listingID); err != nil {
// // 		return err
// // 	}

// // 	return tx.Commit()
// // } nil, domain.ErrInvalidMake
// // 		}
// // 		return nil, err
// // 	}

// // 	return make, nil
// // }

// // func (r *carRepository) GetAllMakes(ctx context.Context) ([]domain.CarMake, error) {
// // 	query := `SELECT id, name, logo_url, country_of_origin, created_at, updated_at FROM car_makes ORDER BY name ASC`

// // 	rows, err := r.db.QueryContext(ctx, query)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	defer rows.Close()

// // 	var makes []domain.CarMake
// // 	for rows.Next() {
// // 		var make domain.CarMake
// // 		err := rows.Scan(&make.ID, &make.Name, &make.LogoURL, &make.CountryOfOrigin, &make.CreatedAt, &make.UpdatedAt)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		makes = append(makes, make)
// // 	}

// // 	return makes, nil
// // }

// // // Model operations

// // func (r *carRepository) GetModelByID(ctx context.Context, id int64) (*domain.CarModel, error) {
// // 	query := `SELECT id, make_id, name, created_at, updated_at FROM car_models WHERE id = ?`

// // 	model := &domain.CarModel{}
// // 	err := r.db.QueryRowContext(ctx, query, id).Scan(
// // 		&model.ID, &model.MakeID, &model.Name, &model.CreatedAt, &model.UpdatedAt,
// // 	)

// // 	if err != nil {
// // 		if errors.Is(err, sql.ErrNoRows) {
// // 			return nil, domain.ErrInvalidModel
// // 		}
// // 		return nil, err
// // 	}

// // 	return model, nil
// // }

// // func (r *carRepository) GetModelsByMakeID(ctx context.Context, makeID int64) ([]domain.CarModel, error) {
// // 	query := `SELECT id, make_id, name, created_at, updated_at FROM car_models WHERE make_id = ? ORDER BY name ASC`

// // 	rows, err := r.db.QueryContext(ctx, query, makeID)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	defer rows.Close()

// // 	var models []domain.CarModel
// // 	for rows.Next() {
// // 		var model domain.CarModel
// // 		err := rows.Scan(&model.ID, &model.MakeID, &model.Name, &model.CreatedAt, &model.UpdatedAt)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		models = append(models, model)
// // 	}

// // 	return models, nil
// // }

// // // Body type operations

// // func (r *carRepository) GetBodyTypeByID(ctx context.Context, id int64) (*domain.BodyType, error) {
// // 	query := `SELECT id, name, created_at FROM body_types WHERE id = ?`

// // 	bodyType := &domain.BodyType{}
// // 	err := r.db.QueryRowContext(ctx, query, id).Scan(&bodyType.ID, &bodyType.Name, &bodyType.CreatedAt)

// // 	if err != nil {
// // 		if errors.Is(err, sql.ErrNoRows) {
// // 			return nil, domain.ErrInvalidBodyType
// // 		}
// // 		return nil, err
// // 	}

// // 	return bodyType, nil
// // }

// // func (r *carRepository) GetAllBodyTypes(ctx context.Context) ([]domain.BodyType, error) {
// // 	query := `SELECT id, name, created_at FROM body_types ORDER BY name ASC`

// // 	rows, err := r.db.QueryContext(ctx, query)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	defer rows.Close()

// // 	var bodyTypes []domain.BodyType
// // 	for rows.Next() {
// // 		var bt domain.BodyType
// // 		err := rows.Scan(&bt.ID, &bt.Name, &bt.CreatedAt)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		bodyTypes = append(bodyTypes, bt)
// // 	}

// // 	return bodyTypes, nil
// // }

// // // Fuel type operations

// // func (r *carRepository) GetFuelTypeByID(ctx context.Context, id int64) (*domain.FuelType, error) {
// // 	query := `SELECT id, name, created_at FROM fuel_types WHERE id = ?`

// // 	fuelType := &domain.FuelType{}
// // 	err := r.db.QueryRowContext(ctx, query, id).Scan(&fuelType.ID, &fuelType.Name, &fuelType.CreatedAt)

// // 	if err != nil {
// // 		if errors.Is(err, sql.ErrNoRows) {
// // 			return nil, domain.ErrInvalidFuelType
// // 		}
// // 		return nil, err
// // 	}

// // 	return fuelType, nil
// // }

// // func (r *carRepository) GetAllFuelTypes(ctx context.Context) ([]domain.FuelType, error) {
// // 	query := `SELECT id, name, created_at FROM fuel_types ORDER BY name ASC`

// // 	rows, err := r.db.QueryContext(ctx, query)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	defer rows.Close()

// // 	var fuelTypes []domain.FuelType
// // 	for rows.Next() {
// // 		var ft domain.FuelType
// // 		err := rows.Scan(&ft.ID, &ft.Name, &ft.CreatedAt)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		fuelTypes = append(fuelTypes, ft)
// // 	}

// // 	return fuelTypes, nil
// // }

// // // Transmission operations

// // func (r *carRepository) GetTransmissionByID(ctx context.Context, id int64) (*domain.TransmissionType, error) {
// // 	query := `SELECT id, name, created_at FROM transmission_types WHERE id = ?`

// // 	transmission := &domain.TransmissionType{}
// // 	err := r.db.QueryRowContext(ctx, query, id).Scan(&transmission.ID, &transmission.Name, &transmission.CreatedAt)

// // 	if err != nil {
// // 		if errors.Is(err, sql.ErrNoRows) {
// // 			return nil, domain.ErrInvalidTransmission
// // 		}
// // 		return nil, err
// // 	}

// // 	return transmission, nil
// // }

// // func (r *carRepository) GetAllTransmissions(ctx context.Context) ([]domain.TransmissionType, error) {
// // 	query := `SELECT id, name, created_at FROM transmission_types ORDER BY name ASC`

// // 	rows, err := r.db.QueryContext(ctx, query)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	defer rows.Close()

// // 	var transmissions []domain.TransmissionType
// // 	for rows.Next() {
// // 		var t domain.TransmissionType
// // 		err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		transmissions = append(transmissions, t)
// // 	}

// // 	return transmissions, nil
// // }

// // // Color operations

// // func (r *carRepository) GetColorByID(ctx context.Context, id int64) (*domain.Color, error) {
// // 	query := `SELECT id, name, hex_code, created_at FROM colors WHERE id = ?`

// // 	color := &domain.Color{}
// // 	err := r.db.QueryRowContext(ctx, query, id).Scan(&color.ID, &color.Name, &color.HexCode, &color.CreatedAt)

// // 	if err != nil {
// // 		if errors.Is(err, sql.ErrNoRows) {
// // 			return nil, domain.ErrInvalidColor
// // 		}
// // 		return nil, err
// // 	}

// // 	return color, nil
// // }

// // func (r *carRepository) GetAllColors(ctx context.Context) ([]domain.Color, error) {
// // 	query := `SELECT id, name, hex_code, created_at FROM colors ORDER BY name ASC`

// // 	rows, err := r.db.QueryContext(ctx, query)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	defer rows.Close()

// // 	var colors []domain.Color
// // 	for rows.Next() {
// // 		var c domain.Color
// // 		err := rows.Scan(&c.ID, &c.Name, &c.HexCode, &c.CreatedAt)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		colors = append(colors, c)
// // 	}

// // 	return colors, nil
// // }

// // // Condition operations

// // func (r *carRepository) GetConditionByID(ctx context.Context, id int64) (*domain.CarCondition, error) {
// // 	query := `SELECT id, name, created_at FROM car_conditions WHERE id = ?`

// // 	condition := &domain.CarCondition{}
// // 	err := r.db.QueryRowContext(ctx, query, id).Scan(&condition.ID, &condition.Name, &condition.CreatedAt)

// // 	if err != nil {
// // 		if errors.Is(err, sql.ErrNoRows) {
// // 			return nil, domain.ErrInvalidCondition
// // 		}
// // 		return nil, err
// // 	}

// // 	return condition, nil
// // }

// // func (r *carRepository) GetAllConditions(ctx context.Context) ([]domain.CarCondition, error) {
// // 	query := `SELECT id, name, created_at FROM car_conditions ORDER BY name ASC`

// // 	rows, err := r.db.QueryContext(ctx, query)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	defer rows.Close()

// // 	var conditions []domain.CarCondition
// // 	for rows.Next() {
// // 		var c domain.CarCondition
// // 		err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		conditions = append(conditions, c)
// // 	}

// // 	return conditions, nil
// // }

// // // Feature operations

// // func (r *carRepository) GetFeatureByID(ctx context.Context, id int64) (*domain.Feature, error) {
// // 	query := `SELECT id, name, category, created_at FROM features WHERE id = ?`

// // 	feature := &domain.Feature{}
// // 	err := r.db.QueryRowContext(ctx, query, id).Scan(&feature.ID, &feature.Name, &feature.Category, &feature.CreatedAt)

// // 	if err != nil {
// // 		if errors.Is(err, sql.ErrNoRows) {
// // 			return nil, domain.ErrInvalidFeature
// // 		}
// // 		return nil, err
// // 	}

// // 	return feature, nil
// // }

// // func (r *carRepository) GetAllFeatures(ctx context.Context) ([]domain.Feature, error) {
// // 	query := `SELECT id, name, category, created_at FROM features ORDER BY category, name ASC`

// // 	rows, err := r.db.QueryContext(ctx, query)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	defer rows.Close()

// // 	var features []domain.Feature
// // 	for rows.Next() {
// // 		var f domain.Feature
// // 		err := rows.Scan(&f.ID, &f.Name, &f.Category, &f.CreatedAt)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		features = append(features, f)
// // 	}

// // 	return features, nil
// // }

// // func (r *carRepository) GetFeaturesByCategory(ctx context.Context, category string) ([]domain.Feature, error) {
// // 	query := `SELECT id, name, category, created_at FROM features WHERE category = ? ORDER BY name ASC`

// // 	rows, err := r.db.QueryContext(ctx, query, category)
// // 	if err != nil {
// // 		return nil, err
// // 	}
// // 	defer rows.Close()

// // 	var features []domain.Feature
// // 	for rows.Next() {
// // 		var f domain.Feature
// // 		err := rows.Scan(&f.ID, &f.Name, &f.Category, &f.CreatedAt)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		features = append(features, f)
// // 	}

// // 	return features, nil
// // }
