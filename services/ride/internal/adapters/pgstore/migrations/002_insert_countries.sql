INSERT INTO countries (code, currency_code, is_enabled) VALUES
-- Eurozone Members
('at', 'EUR', true),  -- Austria
('be', 'EUR', true),  -- Belgium
('bg', 'EUR', true),  -- Bulgaria
('hr', 'EUR', true),  -- Croatia
('cy', 'EUR', true),  -- Cyprus
('ee', 'EUR', true),  -- Estonia
('fi', 'EUR', true),  -- Finland
('fr', 'EUR', true),  -- France
('de', 'EUR', true),  -- Germany
('gr', 'EUR', true),  -- Greece
('ie', 'EUR', true),  -- Ireland
('it', 'EUR', true),  -- Italy
('lv', 'EUR', true),  -- Latvia
('lt', 'EUR', true),  -- Lithuania
('lu', 'EUR', true),  -- Luxembourg
('mt', 'EUR', true),  -- Malta
('nl', 'EUR', true),  -- Netherlands
('pt', 'EUR', true),  -- Portugal
('sk', 'EUR', true),  -- Slovakia
('si', 'EUR', true),  -- Slovenia
('es', 'EUR', true),  -- Spain

-- EU Members (Non-Euro)
('cz', 'CZK', true),  -- Czech Republic
('dk', 'DKK', true),  -- Denmark
('hu', 'HUF', true),  -- Hungary
('pl', 'PLN', true),  -- Poland
('ro', 'RON', true),  -- Romania
('se', 'SEK', true),  -- Sweden

-- Non-EU
('gb', 'GBP', true),  -- United Kingdom
('ch', 'CHF', true),  -- Switzerland
('no', 'NOK', true),  -- Norway
('is', 'ISK', true)   -- Iceland
ON CONFLICT (code) DO UPDATE SET is_enabled = EXCLUDED.is_enabled;