CREATE TABLE daily_stats (
    date DATE NOT NULL,
    language VARCHAR(10) NOT NULL,
    change_count INTEGER NOT NULL DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (date, language)
);