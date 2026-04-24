CREATE TABLE countries (
    country_id SERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    code       CHAR(2) NOT NULL UNIQUE
);

INSERT INTO countries (name, code) VALUES
('Afghanistan', 'AF'), ('Albania', 'AL'), ('Algeria', 'DZ'), ('Argentina', 'AR'),
('Armenia', 'AM'), ('Australia', 'AU'), ('Austria', 'AT'), ('Azerbaijan', 'AZ'),
('Belarus', 'BY'), ('Belgium', 'BE'), ('Bolivia', 'BO'), ('Brazil', 'BR'),
('Bulgaria', 'BG'), ('Canada', 'CA'), ('Chile', 'CL'), ('China', 'CN'),
('Colombia', 'CO'), ('Croatia', 'HR'), ('Cuba', 'CU'), ('Czech Republic', 'CZ'),
('Denmark', 'DK'), ('Ecuador', 'EC'), ('Egypt', 'EG'), ('Estonia', 'EE'),
('Finland', 'FI'), ('France', 'FR'), ('Georgia', 'GE'), ('Germany', 'DE'),
('Ghana', 'GH'), ('Greece', 'GR'), ('Hungary', 'HU'), ('India', 'IN'),
('Indonesia', 'ID'), ('Iran', 'IR'), ('Iraq', 'IQ'), ('Ireland', 'IE'),
('Israel', 'IL'), ('Italy', 'IT'), ('Japan', 'JP'), ('Jordan', 'JO'),
('Kazakhstan', 'KZ'), ('Kenya', 'KE'), ('Kyrgyzstan', 'KG'), ('Latvia', 'LV'),
('Lithuania', 'LT'), ('Malaysia', 'MY'), ('Mexico', 'MX'), ('Moldova', 'MD'),
('Mongolia', 'MN'), ('Morocco', 'MA'), ('Netherlands', 'NL'), ('New Zealand', 'NZ'),
('Nigeria', 'NG'), ('North Korea', 'KP'), ('Norway', 'NO'), ('Pakistan', 'PK'),
('Peru', 'PE'), ('Philippines', 'PH'), ('Poland', 'PL'), ('Portugal', 'PT'),
('Romania', 'RO'), ('Russia', 'RU'), ('Saudi Arabia', 'SA'), ('Serbia', 'RS'),
('Singapore', 'SG'), ('Slovakia', 'SK'), ('Slovenia', 'SI'), ('South Africa', 'ZA'),
('South Korea', 'KR'), ('Spain', 'ES'), ('Sweden', 'SE'), ('Switzerland', 'CH'),
('Taiwan', 'TW'), ('Tajikistan', 'TJ'), ('Thailand', 'TH'), ('Turkey', 'TR'),
('Turkmenistan', 'TM'), ('Ukraine', 'UA'), ('United Kingdom', 'GB'),
('United States', 'US'), ('Uzbekistan', 'UZ'), ('Venezuela', 'VE'), ('Vietnam', 'VN');
