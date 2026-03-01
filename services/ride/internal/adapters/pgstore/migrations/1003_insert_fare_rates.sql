-- Seed Fare Rates for Spain (ES)
INSERT INTO fare_rates (
    country_code, 
    service_type, 
    base_fare, 
    cost_per_km, 
    cost_per_minute, 
    minimum_fare
) VALUES 
('ES', 'STANDARD', 1.50, 1.15, 0.20, 3.50),
('ES', 'PREMIUM', 3.00, 1.80, 0.35, 6.00),
('ES', 'XL', 2.50, 1.60, 0.30, 8.00),
('ES', 'LUXURY', 5.00, 2.00, 0.50, 15.00);