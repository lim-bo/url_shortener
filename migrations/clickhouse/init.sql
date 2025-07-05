CREATE TABLE IF NOT EXISTS urlshortener_stats.redirect_stat (
    link String, 
    code String,
    clicks UInt64
) ENGINE = SummingMergeTree() ORDER BY clicks;