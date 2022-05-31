sudo apt-get install -y golang
sudo apt-get install -y g++ gcc clang 
#install python
sudo apt-get install -y python python3
sudo apt-get install -y python3-pip
#install Mapnik
sudo apt-get install -y git autoconf libtool libxml2-dev libbz2-dev \
libgeos-dev libgeos++-dev libproj-dev gdal-bin libgdal-dev g++ \
libmapnik-dev mapnik-utils python3-mapnik
sudo apt-get install -y libharfbuzz-dev
sudo apt-get install -y mapnik-utils
#install carto and generate XML for Mapnik
git clone https://github.com/gravitystorm/openstreetmap-carto.git
cd openstreetmap-carto
#install fonts
sudo apt-get install -y fonts-noto-cjk fonts-noto-hinted fonts-noto-unhinted fonts-hanazono ttf-unifont
git clone https://github.com/googlefonts/noto-emoji.git
git clone https://github.com/googlefonts/noto-fonts.git
sudo cp noto-emoji/fonts/NotoColorEmoji.ttf /usr/share/fonts/truetype/noto
sudo cp noto-emoji/fonts/NotoEmoji-Regular.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansArabicUI/NotoSansArabicUI-Regular.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoNaskhArabicUI/NotoNaskhArabicUI-Regular.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansArabicUI/NotoSansArabicUI-Bold.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoNaskhArabicUI/NotoNaskhArabicUI-Bold.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansAdlam/NotoSansAdlam-Regular.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansAdlamUnjoined/NotoSansAdlamUnjoined-Regular.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansChakma/NotoSansChakma-Regular.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansOsage/NotoSansOsage-Regular.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansSinhalaUI/NotoSansSinhalaUI-Regular.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansArabicUI/NotoSansArabicUI-Regular.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansCherokee/NotoSansCherokee-Bold.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansSinhalaUI/NotoSansSinhalaUI-Bold.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansSymbols/NotoSansSymbols-Bold.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansArabicUI/NotoSansArabicUI-Bold.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/unhinted/ttf/NotoSansSymbols2/NotoSansSymbols2-Regular.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/hinted/ttf/NotoSansBalinese/NotoSansBalinese-Regular.ttf /usr/share/fonts/truetype/noto
sudo cp noto-fonts/archive/hinted/NotoSansSyriac/NotoSansSyriac-Regular.ttf /usr/share/fonts/truetype/noto
mkdir NotoSansSyriacEastern-unhinted
cd NotoSansSyriacEastern-unhinted
wget https://noto-website-2.storage.googleapis.com/pkgs/NotoSansSyriacEastern-unhinted.zip
unzip NotoSansSyriacEastern-unhinted.zip
sudo cp NotoSansSyriacEastern-Regular.ttf /usr/share/fonts/truetype/noto
cd ..
sudo apt install fontconfig
sudo fc-cache -fv
sudo apt-get install -y fonts-dejavu-core
#install nodejs
sudo apt install -y nodejs npm
#install carto itself
sudo npm install -g carto
npm install mapnik-reference
node -e "console.log(require('mapnik-reference'))"
carto -a "3.0.22" project.mml > style.xml
#install postgres. This is 12 version, but u can always use more old versions
sudo apt-get install -y postgresql postgis postgresql-contrib postgresql-postgis postgresql-postgis-scripts
sudo service postgresql start
#configure postgres database
sudo -u postgres createuser -s $USER
sudo -u postgres createdb gis
psql -d gis -c 'CREATE EXTENSION postgis; CREATE EXTENSION hstore;'
psql -d gis -c 'ALTER TABLE geometry_columns OWNER TO postgres;'
psql -d gis -c 'ALTER TABLE spatial_ref_sys OWNER TO  postgres;'
#install osm2pqsql for importing osm data
sudo apt install -y osm2pgsql
#download osm map
wget -c http://download.geofabrik.de/russia/central-fed-district-latest.osm.pbf
osm2pgsql \
--number-processes 16 \
--hstore \
--multi-geometry \
--database gis \
--slim \
--drop \
-U $USER \
--style openstreetmap-carto.style \
--tag-transform-script openstreetmap-carto.lua \
central-fed-district-latest.osm.pbf
#now we need to generate 'data' folder, that Mapnik will use for render
python3 -m pip install psycopg2-binary
#maybe you'll have to give permissions to user write to 'data'
scripts/get-external-data.py
#now we generate indexes for our database
HOSTNAME=localhost
psql -d gis -f indexes.sql
#now we are ready to go. We need to copy 'data', 'style', 'symbols', 'patterns' and style.xml to our server folder
