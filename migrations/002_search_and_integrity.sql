-- ============================================
-- Migration: 002_search_and_integrity.sql
-- Full-text search, index fixes and integrity rules
-- ============================================

-- Full-text search over listing title and description. An external-content
-- FTS5 table indexes car_listings without duplicating the text; triggers
-- keep it in sync.
CREATE VIRTUAL TABLE car_listings_fts USING fts5(
    title,
    description,
    content='car_listings',
    content_rowid='id'
);

INSERT INTO car_listings_fts(car_listings_fts) VALUES ('rebuild');

CREATE TRIGGER car_listings_fts_ai AFTER INSERT ON car_listings BEGIN
    INSERT INTO car_listings_fts(rowid, title, description)
    VALUES (new.id, new.title, new.description);
END;

CREATE TRIGGER car_listings_fts_ad AFTER DELETE ON car_listings BEGIN
    INSERT INTO car_listings_fts(car_listings_fts, rowid, title, description)
    VALUES ('delete', old.id, old.title, old.description);
END;

CREATE TRIGGER car_listings_fts_au AFTER UPDATE OF title, description ON car_listings BEGIN
    INSERT INTO car_listings_fts(car_listings_fts, rowid, title, description)
    VALUES ('delete', old.id, old.title, old.description);
    INSERT INTO car_listings_fts(rowid, title, description)
    VALUES (new.id, new.title, new.description);
END;

-- A missing VIN is NULL, not '', so the UNIQUE constraint only applies to real VINs.
UPDATE car_listings SET vin = NULL WHERE vin = '';

-- Location filters are case-insensitive; the index must use the same collation.
DROP INDEX IF EXISTS idx_car_listings_location;
CREATE INDEX idx_car_listings_location
    ON car_listings(city COLLATE NOCASE, state COLLATE NOCASE, country COLLATE NOCASE);

-- These duplicate the automatic indexes created by their UNIQUE constraints.
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_refresh_tokens_token;

CREATE INDEX IF NOT EXISTS idx_car_images_listing ON car_images(car_listing_id, display_order);

-- UNIQUE(reviewer_id, reviewed_user_id, car_listing_id) treats NULLs as
-- distinct, so it never stopped duplicate general (listing-less) reviews.
CREATE UNIQUE INDEX idx_reviews_unique_general
    ON reviews(reviewer_id, reviewed_user_id)
    WHERE car_listing_id IS NULL;

-- Refresh tokens are now stored as SHA-256 hashes; existing plaintext
-- tokens can no longer be matched, so users simply log in again.
DELETE FROM refresh_tokens;
