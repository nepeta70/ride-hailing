INSERT INTO countries (code, currency_code, is_enabled) VALUES
-- Eurozone Members
('AT', 'EUR', true),  -- Austria
('BE', 'EUR', true),  -- Belgium
('BG', 'EUR', true),  -- Bulgaria
('HR', 'EUR', true),  -- Croatia
('CY', 'EUR', true),  -- Cyprus
('EE', 'EUR', true),  -- Estonia
('FI', 'EUR', true),  -- Finland
('FR', 'EUR', true),  -- France
('DE', 'EUR', true),  -- Germany
('GR', 'EUR', true),  -- Greece
('IE', 'EUR', true),  -- Ireland
('IT', 'EUR', true),  -- Italy
('LV', 'EUR', true),  -- Latvia
('LT', 'EUR', true),  -- Lithuania
('LU', 'EUR', true),  -- Luxembourg
('MT', 'EUR', true),  -- Malta
('NL', 'EUR', true),  -- Netherlands
('PT', 'EUR', true),  -- Portugal
('SK', 'EUR', true),  -- Slovakia
('SI', 'EUR', true),  -- Slovenia
('ES', 'EUR', true),  -- Spain

-- EU Members (Non-Euro)
('CZ', 'CZK', true),  -- Czech Republic
('DK', 'DKK', true),  -- Denmark
('HU', 'HUF', true),  -- Hungary
('PL', 'PLN', true),  -- Poland
('RO', 'RON', true),  -- Romania
('SE', 'SEK', true),  -- Sweden

-- Non-EU
('GB', 'GBP', true),  -- United Kingdom
('CH', 'CHF', true),  -- Switzerland
('NO', 'NOK', true),  -- Norway
('IS', 'ISK', true)   -- Iceland
ON CONFLICT (code) DO UPDATE SET is_enabled = EXCLUDED.is_enabled;