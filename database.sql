CREATE DATABASE map_db;
CREATE USER map_app WITH ENCRYPTED PASSWORD '12345';
\c map_db;
CREATE EXTENSION postgis;
CREATE EXTENSION postgis_raster;
SET postgis.enable_outdb_rasters TO True;
SET postgis.gdal_enabled_drivers TO 'PNG';

SELECT pg_reload_conf();

GRANT ALL PRIVILEGES ON DATABASE map_db to map_app;
GRANT USAGE ON SCHEMA public TO map_app;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO map_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO map_app;