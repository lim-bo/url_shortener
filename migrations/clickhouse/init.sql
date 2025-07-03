CREATE TABLE IF NOT EXISTS default.redirect_stat (
    link String, 
    code String,
    clicks UInt64
) ENGINE = MergeTree() ORDER BY clicks;