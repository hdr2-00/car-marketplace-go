-- ============================================
-- Sample Data for Testing
-- Run this after initial schema setup
-- ============================================

-- Insert Car Makes
INSERT INTO car_makes (name, country_of_origin, created_at, updated_at) VALUES
('Toyota', 'Japan', datetime('now'), datetime('now')),
('Honda', 'Japan', datetime('now'), datetime('now')),
('Ford', 'USA', datetime('now'), datetime('now')),
('BMW', 'Germany', datetime('now'), datetime('now')),
('Mercedes-Benz', 'Germany', datetime('now'), datetime('now')),
('Volkswagen', 'Germany', datetime('now'), datetime('now')),
('Audi', 'Germany', datetime('now'), datetime('now')),
('Nissan', 'Japan', datetime('now'), datetime('now')),
('Chevrolet', 'USA', datetime('now'), datetime('now')),
('Hyundai', 'South Korea', datetime('now'), datetime('now')),
('Kia', 'South Korea', datetime('now'), datetime('now')),
('Tesla', 'USA', datetime('now'), datetime('now')),
('Mazda', 'Japan', datetime('now'), datetime('now')),
('Subaru', 'Japan', datetime('now'), datetime('now')),
('Lexus', 'Japan', datetime('now'), datetime('now'));

-- Insert Car Models for Toyota
INSERT INTO car_models (make_id, name, created_at, updated_at) VALUES
(1, 'Camry', datetime('now'), datetime('now')),
(1, 'Corolla', datetime('now'), datetime('now')),
(1, 'RAV4', datetime('now'), datetime('now')),
(1, 'Highlander', datetime('now'), datetime('now')),
(1, 'Prius', datetime('now'), datetime('now'));

-- Insert Car Models for Honda
INSERT INTO car_models (make_id, name, created_at, updated_at) VALUES
(2, 'Accord', datetime('now'), datetime('now')),
(2, 'Civic', datetime('now'), datetime('now')),
(2, 'CR-V', datetime('now'), datetime('now')),
(2, 'Pilot', datetime('now'), datetime('now')),
(2, 'Odyssey', datetime('now'), datetime('now'));

-- Insert Car Models for Ford
INSERT INTO car_models (make_id, name, created_at, updated_at) VALUES
(3, 'F-150', datetime('now'), datetime('now')),
(3, 'Mustang', datetime('now'), datetime('now')),
(3, 'Explorer', datetime('now'), datetime('now')),
(3, 'Escape', datetime('now'), datetime('now')),
(3, 'Focus', datetime('now'), datetime('now'));

-- Insert Car Models for BMW
INSERT INTO car_models (make_id, name, created_at, updated_at) VALUES
(4, '3 Series', datetime('now'), datetime('now')),
(4, '5 Series', datetime('now'), datetime('now')),
(4, 'X3', datetime('now'), datetime('now')),
(4, 'X5', datetime('now'), datetime('now')),
(4, 'M3', datetime('now'), datetime('now'));

-- Insert Car Models for Mercedes-Benz
INSERT INTO car_models (make_id, name, created_at, updated_at) VALUES
(5, 'C-Class', datetime('now'), datetime('now')),
(5, 'E-Class', datetime('now'), datetime('now')),
(5, 'GLC', datetime('now'), datetime('now')),
(5, 'GLE', datetime('now'), datetime('now')),
(5, 'S-Class', datetime('now'), datetime('now'));

-- Insert Car Models for Tesla
INSERT INTO car_models (make_id, name, created_at, updated_at) VALUES
(12, 'Model 3', datetime('now'), datetime('now')),
(12, 'Model Y', datetime('now'), datetime('now')),
(12, 'Model S', datetime('now'), datetime('now')),
(12, 'Model X', datetime('now'), datetime('now'));

-- Sample Test Users
-- Password for all users: "password123" (hashed with bcrypt cost 10)
-- Hash: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy

INSERT INTO users (email, password_hash, user_type_id, is_verified, is_active, created_at, updated_at) VALUES
('john.doe@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1, 1, 1, datetime('now'), datetime('now')),
('jane.smith@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 1, 1, 1, datetime('now'), datetime('now')),
('premium.motors@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 2, 1, 1, datetime('now'), datetime('now')),
('elite.showroom@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 3, 1, 1, datetime('now'), datetime('now'));

-- Individual User Profiles
INSERT INTO user_profiles (user_id, first_name, last_name, phone_number, created_at, updated_at) VALUES
(1, 'John', 'Doe', '+1234567890', datetime('now'), datetime('now')),
(2, 'Jane', 'Smith', '+1234567891', datetime('now'), datetime('now'));

-- Business Profiles
INSERT INTO business_profiles (user_id, business_name, phone_number, description, established_year, created_at, updated_at) VALUES
(3, 'Premium Motors', '+1234567892', 'Your trusted car dealership since 2005', 2005, datetime('now'), datetime('now')),
(4, 'Elite Auto Showroom', '+1234567893', 'Luxury and premium vehicles', 2010, datetime('now'), datetime('now'));

-- Sample Car Listings
INSERT INTO car_listings (
    seller_id, make_id, model_id, year, body_type_id, fuel_type_id,
    transmission_type_id, color_id, condition_id, mileage, engine_size,
    horsepower, number_of_doors, seating_capacity, price, currency, is_negotiable,
    title, description, city, state, country, status, is_featured, views_count,
    created_at, updated_at
) VALUES
-- John's car
(1, 1, 1, 2020, 1, 1, 2, 1, 2, 25000, 2.5, 203, 4, 5, 24500.00, 'USD', 1,
 '2020 Toyota Camry - Excellent Condition',
 'Well-maintained Toyota Camry with full service history. Single owner, garage kept. Features include leather seats, navigation, and backup camera.',
 'New York', 'NY', 'USA', 'active', 0, 45, datetime('now', '-5 days'), datetime('now', '-5 days')),

-- Jane's car
(2, 2, 2, 2019, 2, 1, 2, 3, 2, 32000, 1.5, 158, 4, 5, 18900.00, 'USD', 1,
 '2019 Honda Civic - Low Mileage',
 'Great fuel economy and reliable transportation. Recent oil change and new tires. Clean title, no accidents.',
 'Los Angeles', 'CA', 'USA', 'active', 0, 32, datetime('now', '-3 days'), datetime('now', '-3 days')),

-- Premium Motors listing 1
(3, 4, 1, 2021, 1, 2, 2, 2, 3, 15000, 2.0, 255, 4, 5, 38900.00, 'USD', 1,
 '2021 BMW 3 Series - Certified Pre-Owned',
 'BMW Certified Pre-Owned with extended warranty. Premium package, sport seats, and advanced safety features.',
 'Chicago', 'IL', 'USA', 'active', 1, 128, datetime('now', '-7 days'), datetime('now', '-7 days')),

-- Premium Motors listing 2
(3, 5, 1, 2022, 1, 2, 2, 1, 3, 8000, 2.0, 255, 4, 5, 45900.00, 'USD', 1,
 '2022 Mercedes-Benz C-Class - Like New',
 'Barely driven Mercedes-Benz C-Class. Premium interior, panoramic sunroof, advanced driver assistance.',
 'Chicago', 'IL', 'USA', 'active', 1, 95, datetime('now', '-6 days'), datetime('now', '-6 days')),

-- Elite Showroom listing
(4, 12, 1, 2023, 1, 4, 2, 4, 1, 5000, 0.0, 283, 4, 5, 42900.00, 'USD', 0,
 '2023 Tesla Model 3 - New Condition',
 'Nearly new Tesla Model 3 Long Range. Autopilot, premium connectivity, all the latest features.',
 'Miami', 'FL', 'USA', 'active', 1, 156, datetime('now', '-2 days'), datetime('now', '-2 days'));

-- Add features to some listings
-- BMW features (listing 3)
INSERT INTO car_listing_features (car_listing_id, feature_id, created_at) VALUES
(3, 1, datetime('now')), -- abs
(3, 2, datetime('now')), -- airbags
(3, 3, datetime('now')), -- backup_camera
(3, 7, datetime('now')), -- leather_seats
(3, 11, datetime('now')), -- sunroof
(3, 13, datetime('now')), -- climate_control
(3, 14, datetime('now')), -- navigation_system
(3, 15, datetime('now')); -- bluetooth

-- Mercedes features (listing 4)
INSERT INTO car_listing_features (car_listing_id, feature_id, created_at) VALUES
(4, 1, datetime('now')), -- abs
(4, 2, datetime('now')), -- airbags
(4, 3, datetime('now')), -- backup_camera
(4, 4, datetime('now')), -- blind_spot_monitoring
(4, 7, datetime('now')), -- leather_seats
(4, 8, datetime('now')), -- heated_seats
(4, 12, datetime('now')), -- panoramic_roof
(4, 13, datetime('now')), -- climate_control
(4, 14, datetime('now')), -- navigation_system
(4, 16, datetime('now')); -- premium_audio

-- Tesla features (listing 5)
INSERT INTO car_listing_features (car_listing_id, feature_id, created_at) VALUES
(5, 2, datetime('now')), -- airbags
(5, 3, datetime('now')), -- backup_camera
(5, 4, datetime('now')), -- blind_spot_monitoring
(5, 5, datetime('now')), -- lane_departure_warning
(5, 8, datetime('now')), -- heated_seats
(5, 14, datetime('now')), -- navigation_system
(5, 15, datetime('now')), -- bluetooth
(5, 17, datetime('now')), -- apple_carplay
(5, 22, datetime('now')), -- adaptive_cruise_control
(5, 24, datetime('now')), -- keyless_entry
(5, 25, datetime('now')); -- push_start

-- Sample Car Images
INSERT INTO car_images (car_listing_id, image_url, display_order, is_primary, created_at) VALUES
(1, 'https://example.com/images/toyota-camry-front.jpg', 0, 1, datetime('now')),
(1, 'https://example.com/images/toyota-camry-side.jpg', 1, 0, datetime('now')),
(1, 'https://example.com/images/toyota-camry-interior.jpg', 2, 0, datetime('now')),
(2, 'https://example.com/images/honda-civic-front.jpg', 0, 1, datetime('now')),
(2, 'https://example.com/images/honda-civic-side.jpg', 1, 0, datetime('now')),
(3, 'https://example.com/images/bmw-3series-front.jpg', 0, 1, datetime('now')),
(3, 'https://example.com/images/bmw-3series-interior.jpg', 1, 0, datetime('now')),
(4, 'https://example.com/images/mercedes-cclass-front.jpg', 0, 1, datetime('now')),
(5, 'https://example.com/images/tesla-model3-front.jpg', 0, 1, datetime('now')),
(5, 'https://example.com/images/tesla-model3-interior.jpg', 1, 0, datetime('now'));

-- Sample Favorites
INSERT INTO favorites (user_id, car_listing_id, created_at) VALUES
(1, 3, datetime('now', '-2 days')),
(1, 5, datetime('now', '-1 day')),
(2, 3, datetime('now', '-3 days')),
(2, 4, datetime('now', '-2 days'));

-- Sample Inquiries
INSERT INTO inquiries (car_listing_id, buyer_id, seller_id, subject, message, is_read, created_at) VALUES
(3, 1, 3, 'Interested in BMW', 'Hi, I am interested in the BMW 3 Series. Is it still available?', 1, datetime('now', '-1 day')),
(5, 2, 4, 'Tesla Model 3 Question', 'Does this Tesla have the enhanced autopilot feature?', 0, datetime('now', '-6 hours')),
(1, 2, 1, 'Toyota Camry Inquiry', 'Can you provide more photos of the interior?', 1, datetime('now', '-2 days'));

-- Sample Reviews
INSERT INTO reviews (reviewer_id, reviewed_user_id, car_listing_id, rating, title, comment, is_verified_purchase, created_at, updated_at) VALUES
(1, 3, 3, 5, 'Excellent Service!', 'Premium Motors was very professional. The car was exactly as described and the buying process was smooth.', 0, datetime('now', '-1 day'), datetime('now', '-1 day')),
(2, 3, NULL, 5, 'Great Dealership', 'Highly recommend Premium Motors. They have quality cars and honest staff.', 0, datetime('now', '-3 days'), datetime('now', '-3 days')),
(1, 4, 5, 4, 'Good Experience', 'Elite Showroom was helpful and knowledgeable about the Tesla. Minor wait time but overall good.', 0, datetime('now', '-12 hours'), datetime('now', '-12 hours'));