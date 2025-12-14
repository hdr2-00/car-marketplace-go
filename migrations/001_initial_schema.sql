-- ============================================
-- Car Marketplace Database Schema
-- SQLite3 with Normalization
-- Migration: 001_initial_schema.sql
-- ============================================

-- ============================================
-- User Management Tables
-- ============================================

-- User Types/Roles
CREATE TABLE user_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE, -- 'individual', 'dealership', 'showroom'
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Users Table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    user_type_id INTEGER NOT NULL,
    is_verified BOOLEAN DEFAULT 0,
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_type_id) REFERENCES user_types(id)
);

-- User Profiles
CREATE TABLE user_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone_number VARCHAR(20),
    profile_image_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Dealership/Showroom Profiles
CREATE TABLE business_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    business_name VARCHAR(255) NOT NULL,
    business_license VARCHAR(100),
    tax_id VARCHAR(100),
    phone_number VARCHAR(20),
    website_url TEXT,
    logo_url TEXT,
    description TEXT,
    established_year INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Business Addresses
CREATE TABLE business_addresses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    business_profile_id INTEGER NOT NULL,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    is_primary BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (business_profile_id) REFERENCES business_profiles(id) ON DELETE CASCADE
);

-- ============================================
-- Authentication Tables
-- ============================================

-- Refresh Tokens
CREATE TABLE refresh_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token VARCHAR(500) NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    is_revoked BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- ============================================
-- Car Reference Data Tables
-- ============================================

-- Car Makes
CREATE TABLE car_makes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE,
    logo_url TEXT,
    country_of_origin VARCHAR(100),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Car Models
CREATE TABLE car_models (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    make_id INTEGER NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (make_id) REFERENCES car_makes(id),
    UNIQUE(make_id, name)
);

-- Body Types
CREATE TABLE body_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE, -- 'sedan', 'suv', 'coupe', 'hatchback', etc.
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Fuel Types
CREATE TABLE fuel_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE, -- 'petrol', 'diesel', 'electric', 'hybrid', etc.
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Transmission Types
CREATE TABLE transmission_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE, -- 'manual', 'automatic', 'cvt', etc.
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Colors
CREATE TABLE colors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE,
    hex_code VARCHAR(7),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Car Conditions
CREATE TABLE car_conditions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE, -- 'new', 'used', 'certified_pre_owned'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- Car Listings Tables
-- ============================================

-- Car Listings
CREATE TABLE car_listings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    seller_id INTEGER NOT NULL,
    make_id INTEGER NOT NULL,
    model_id INTEGER NOT NULL,
    year INTEGER NOT NULL,
    body_type_id INTEGER NOT NULL,
    fuel_type_id INTEGER NOT NULL,
    transmission_type_id INTEGER NOT NULL,
    color_id INTEGER NOT NULL,
    condition_id INTEGER NOT NULL,
    
    -- Car Details
    vin VARCHAR(17) UNIQUE, -- Vehicle Identification Number
    mileage INTEGER NOT NULL,
    engine_size DECIMAL(4, 2), -- in liters (e.g., 2.0L)
    horsepower INTEGER,
    number_of_doors INTEGER,
    seating_capacity INTEGER,
    
    -- Pricing
    price DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    is_negotiable BOOLEAN DEFAULT 1,
    
    -- Listing Details
    title VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Location
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20),
    
    -- Status
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'sold', 'pending', 'expired', 'inactive'
    is_featured BOOLEAN DEFAULT 0,
    views_count INTEGER DEFAULT 0,
    
    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    sold_at DATETIME,
    
    FOREIGN KEY (seller_id) REFERENCES users(id),
    FOREIGN KEY (make_id) REFERENCES car_makes(id),
    FOREIGN KEY (model_id) REFERENCES car_models(id),
    FOREIGN KEY (body_type_id) REFERENCES body_types(id),
    FOREIGN KEY (fuel_type_id) REFERENCES fuel_types(id),
    FOREIGN KEY (transmission_type_id) REFERENCES transmission_types(id),
    FOREIGN KEY (color_id) REFERENCES colors(id),
    FOREIGN KEY (condition_id) REFERENCES car_conditions(id)
);

-- Car Features (Many-to-Many relationship)
CREATE TABLE features (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(100) NOT NULL UNIQUE, -- 'sunroof', 'leather_seats', 'navigation', etc.
    category VARCHAR(50), -- 'safety', 'comfort', 'entertainment', 'technology'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE car_listing_features (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    car_listing_id INTEGER NOT NULL,
    feature_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (car_listing_id) REFERENCES car_listings(id) ON DELETE CASCADE,
    FOREIGN KEY (feature_id) REFERENCES features(id),
    UNIQUE(car_listing_id, feature_id)
);

-- Car Images
CREATE TABLE car_images (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    car_listing_id INTEGER NOT NULL,
    image_url TEXT NOT NULL,
    display_order INTEGER DEFAULT 0,
    is_primary BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (car_listing_id) REFERENCES car_listings(id) ON DELETE CASCADE
);

-- ============================================
-- Transaction Tables
-- ============================================

-- Inquiries/Messages
CREATE TABLE inquiries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    car_listing_id INTEGER NOT NULL,
    buyer_id INTEGER NOT NULL,
    seller_id INTEGER NOT NULL,
    subject VARCHAR(255),
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (car_listing_id) REFERENCES car_listings(id) ON DELETE CASCADE,
    FOREIGN KEY (buyer_id) REFERENCES users(id),
    FOREIGN KEY (seller_id) REFERENCES users(id)
);

-- Favorites/Wishlist
CREATE TABLE favorites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    car_listing_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (car_listing_id) REFERENCES car_listings(id) ON DELETE CASCADE,
    UNIQUE(user_id, car_listing_id)
);

-- Reviews/Ratings
CREATE TABLE reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    reviewer_id INTEGER NOT NULL,
    reviewed_user_id INTEGER NOT NULL,
    car_listing_id INTEGER,
    rating INTEGER NOT NULL CHECK(rating >= 1 AND rating <= 5),
    title VARCHAR(255),
    comment TEXT,
    is_verified_purchase BOOLEAN DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (reviewer_id) REFERENCES users(id),
    FOREIGN KEY (reviewed_user_id) REFERENCES users(id),
    FOREIGN KEY (car_listing_id) REFERENCES car_listings(id) ON DELETE SET NULL,
    UNIQUE(reviewer_id, reviewed_user_id, car_listing_id)
);

-- ============================================
-- Indexes for Performance
-- ============================================

-- Users
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_user_type ON users(user_type_id);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Refresh Tokens
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Car Listings
CREATE INDEX idx_car_listings_seller_id ON car_listings(seller_id);
CREATE INDEX idx_car_listings_make_model ON car_listings(make_id, model_id);
CREATE INDEX idx_car_listings_status ON car_listings(status);
CREATE INDEX idx_car_listings_price ON car_listings(price);
CREATE INDEX idx_car_listings_year ON car_listings(year);
CREATE INDEX idx_car_listings_location ON car_listings(city, state, country);
CREATE INDEX idx_car_listings_created_at ON car_listings(created_at);

-- Favorites
CREATE INDEX idx_favorites_user_id ON favorites(user_id);
CREATE INDEX idx_favorites_car_listing_id ON favorites(car_listing_id);

-- Inquiries
CREATE INDEX idx_inquiries_buyer_id ON inquiries(buyer_id);
CREATE INDEX idx_inquiries_seller_id ON inquiries(seller_id);
CREATE INDEX idx_inquiries_car_listing_id ON inquiries(car_listing_id);

-- Reviews
CREATE INDEX idx_reviews_reviewed_user_id ON reviews(reviewed_user_id);
CREATE INDEX idx_reviews_reviewer_id ON reviews(reviewer_id);

-- ============================================
-- Initial Reference Data
-- ============================================

-- Insert User Types
INSERT INTO user_types (name, description) VALUES
('individual', 'Individual car owner'),
('dealership', 'Licensed car dealership'),
('showroom', 'Car showroom or retail outlet');

-- Insert Body Types
INSERT INTO body_types (name) VALUES
('sedan'), ('suv'), ('hatchback'), ('coupe'), ('convertible'),
('wagon'), ('van'), ('truck'), ('minivan'), ('crossover');

-- Insert Fuel Types
INSERT INTO fuel_types (name) VALUES
('petrol'), ('diesel'), ('electric'), ('hybrid'),
('plug_in_hybrid'), ('cng'), ('lpg');

-- Insert Transmission Types
INSERT INTO transmission_types (name) VALUES
('manual'), ('automatic'), ('cvt'), ('dct'), ('amt');

-- Insert Car Conditions
INSERT INTO car_conditions (name) VALUES
('new'), ('used'), ('certified_pre_owned');

-- Insert Common Colors
INSERT INTO colors (name, hex_code) VALUES
('black', '#000000'),
('white', '#FFFFFF'),
('silver', '#C0C0C0'),
('gray', '#808080'),
('red', '#FF0000'),
('blue', '#0000FF'),
('green', '#008000'),
('yellow', '#FFFF00'),
('orange', '#FFA500'),
('brown', '#A52A2A'),
('beige', '#F5F5DC'),
('gold', '#FFD700');

-- Insert Common Features
INSERT INTO features (name, category) VALUES
-- Safety
('abs', 'safety'),
('airbags', 'safety'),
('backup_camera', 'safety'),
('blind_spot_monitoring', 'safety'),
('lane_departure_warning', 'safety'),
('traction_control', 'safety'),
-- Comfort
('leather_seats', 'comfort'),
('heated_seats', 'comfort'),
('cooled_seats', 'comfort'),
('power_seats', 'comfort'),
('sunroof', 'comfort'),
('panoramic_roof', 'comfort'),
('climate_control', 'comfort'),
-- Entertainment
('navigation_system', 'entertainment'),
('bluetooth', 'entertainment'),
('premium_audio', 'entertainment'),
('apple_carplay', 'entertainment'),
('android_auto', 'entertainment'),
-- Technology
('cruise_control', 'technology'),
('adaptive_cruise_control', 'technology'),
('parking_sensors', 'technology'),
('keyless_entry', 'technology'),
('push_start', 'technology'),
('wireless_charging', 'technology');