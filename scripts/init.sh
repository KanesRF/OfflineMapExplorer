service postgresql start
cd /openstreetmap-carto
sudo -u postgres createuser -s root
sudo -u postgres createdb gis
psql -h localhost -d gis -c 'CREATE EXTENSION postgis; CREATE EXTENSION hstore;'
psql -h localhost -d gis -c 'ALTER TABLE geometry_columns OWNER TO postgres;'
psql -h localhost -d gis -c 'ALTER TABLE spatial_ref_sys OWNER TO  postgres;'
scripts/get-external-data.py
cp -R data /OfflineMapExplorer/data
cp -R symbols /OfflineMapExplorer/symbols
cp -R style /OfflineMapExplorer/style
cp -R patterns /OfflineMapExplorer/patterns
cp style.xml /OfflineMapExplorer/style.xml
osm2pgsql \
--number-processes 16 \
--hstore \
--multi-geometry \
--database gis \
--slim \
--drop \
-U root \
-H 127.0.0.1 \
--style openstreetmap-carto.style \
--tag-transform-script openstreetmap-carto.lua \
central-fed-district-latest.osm.pbf
ENV HOSTNAME=localhost
psql -h localhost -d gis -f indexes.sql
cd /OfflineMapExplorer
make all