service postgresql start
cd /openstreetmap-carto
sudo -u postgres createuser -s root
sudo -u postgres createdb gis
psql -d gis -c 'CREATE EXTENSION postgis; CREATE EXTENSION hstore;'
psql -d gis -c 'ALTER TABLE geometry_columns OWNER TO postgres;'
psql -d gis -c 'ALTER TABLE spatial_ref_sys OWNER TO  postgres;'
osm2pgsql \
--number-processes 16 \
--hstore \
--multi-geometry \
--database gis \
--slim \
--drop \
--style openstreetmap-carto.style \
--tag-transform-script openstreetmap-carto.lua \
central-fed-district-latest.osm.pbf
scripts/get-external-data.py
psql -d gis -f indexes.sql
cp -R data /OfflineMapExplorer/
cp -R symbols /OfflineMapExplorer/
cp -R style /OfflineMapExplorer/
cp -R patterns /OfflineMapExplorer/
cp style.xml /OfflineMapExplorer/
cd /OfflineMapExplorer
make all