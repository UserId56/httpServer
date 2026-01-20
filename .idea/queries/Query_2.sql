SELECT name, abbrev, utc_offset FROM pg_timezone_names
WHERE name = current_setting('TimeZone');