INSERT INTO service_types (
    id, 
    display_name, 
    max_passengers, 
    sort_order
)
VALUES 
    ('STANDARD', 'Standard Ride', 4, 10),
    ('PREMIUM',  'Premium Sedan', 3, 20),
    ('XL',       'Extra Large / Van', 6, 30),
    ('LUXURY',   'Luxury Black', 3, 40)
ON CONFLICT (id) 
DO UPDATE SET 
    display_name    = EXCLUDED.display_name,
    max_passengers  = EXCLUDED.max_passengers,
    sort_order      = EXCLUDED.sort_order,
    current_version = service_types.current_version + 1;