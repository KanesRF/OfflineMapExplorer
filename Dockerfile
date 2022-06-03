FROM ubuntu:20.04
ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
RUN apt-get update
RUN apt-get install -y golang g++ gcc clang python3 python3-pip npm \ 
autoconf libtool libxml2-dev libbz2-dev \
libgeos-dev libgeos++-dev libproj-dev gdal-bin libgdal-dev g++ \
libmapnik-dev mapnik-utils python3-mapnik fonts-dejavu-core fontconfig \ 
postgresql postgis postgresql-contrib postgresql-postgis postgresql-postgis-scripts  \ 
osm2pgsql wget unzip gosu

RUN wget -c https://codeload.github.com/KanesRF/OfflineMapExplorer/zip/refs/heads/master -O OfflineMapExplorer.zip && unzip OfflineMapExplorer.zip && rm OfflineMapExplorer.zip && mv $(ls -d OfflineMapExplorer*) OfflineMapExplorer
RUN wget -c https://codeload.github.com/gravitystorm/openstreetmap-carto/zip/refs/tags/v5.4.0 -O openstreetmap-carto.zip && unzip openstreetmap-carto.zip && rm openstreetmap-carto.zip && mv $(ls -d openstreetmap-carto*) openstreetmap-carto
WORKDIR /openstreetmap-carto
RUN apt-get install -y fonts-noto-cjk fonts-noto-hinted fonts-noto-unhinted fonts-hanazono ttf-unifont
RUN wget -c https://codeload.github.com/googlefonts/noto-emoji/zip/refs/tags/v2.028 -O noto-emoji.zip && unzip noto-emoji.zip && rm noto-emoji.zip && mv $(ls -d noto-emoji*) noto-emoji
RUN wget -c https://codeload.github.com/googlefonts/noto-fonts/zip/refs/tags/v20201206-phase3 -O noto-fonts.zip && unzip noto-fonts.zip && rm noto-fonts.zip && mv $(ls -d noto-fonts*) noto-fonts
RUN cp noto-emoji/fonts/NotoColorEmoji.ttf \
    noto-fonts/hinted/ttf/NotoNaskhArabicUI/NotoNaskhArabicUI-Regular.ttf \
    noto-fonts/hinted/ttf/NotoSansArabicUI/NotoSansArabicUI-Regular.ttf \
    noto-fonts/hinted/ttf/NotoSansArabicUI/NotoSansArabicUI-Bold.ttf \ 
    noto-fonts/hinted/ttf/NotoNaskhArabicUI/NotoNaskhArabicUI-Bold.ttf \ 
    noto-fonts/hinted/ttf/NotoSansAdlam/NotoSansAdlam-Regular.ttf \ 
    noto-fonts/hinted/ttf/NotoSansAdlamUnjoined/NotoSansAdlamUnjoined-Regular.ttf \ 
    noto-fonts/hinted/ttf/NotoSansChakma/NotoSansChakma-Regular.ttf \ 
    noto-fonts/hinted/ttf/NotoSansOsage/NotoSansOsage-Regular.ttf \ 
    noto-fonts/hinted/ttf/NotoSansSinhalaUI/NotoSansSinhalaUI-Regular.ttf \ 
    noto-fonts/hinted/ttf/NotoSansCherokee/NotoSansCherokee-Bold.ttf \ 
    noto-fonts/hinted/ttf/NotoSansSinhalaUI/NotoSansSinhalaUI-Bold.ttf \ 
    noto-fonts/hinted/ttf/NotoSansSymbols/NotoSansSymbols-Bold.ttf \ 
    noto-fonts/unhinted/ttf/NotoSansSymbols2/NotoSansSymbols2-Regular.ttf \
    noto-fonts/hinted/ttf/NotoSansBalinese/NotoSansBalinese-Regular.ttf \  
    noto-fonts/archive/hinted/NotoSansSyriac/NotoSansSyriac-Regular.ttf \ 
    /usr/share/fonts/truetype/noto
RUN mkdir NotoSansSyriacEastern-unhinted
WORKDIR /openstreetmap-carto/NotoSansSyriacEastern-unhinted
RUN wget https://noto-website-2.storage.googleapis.com/pkgs/NotoSansSyriacEastern-unhinted.zip && unzip NotoSansSyriacEastern-unhinted.zip && rm NotoSansSyriacEastern-unhinted.zip
RUN cp NotoSansSyriacEastern-Regular.ttf /usr/share/fonts/truetype/noto
WORKDIR /openstreetmap-carto
RUN fc-cache -fv
RUN npm install -g carto && npm install mapnik-reference
RUN carto -a "3.0.22" project.mml > style.xml
RUN wget -c http://download.geofabrik.de/russia/central-fed-district-latest.osm.pbf
RUN python3 -m pip install psycopg2-binary requests
RUN apt-get install -y sudo
RUN sudo /bin/bash /OfflineMapExplorer/scripts/init.sh
WORKDIR /OfflineMapExplorer/js
RUN wget -c https://github.com/Leaflet/Leaflet/releases/download/v1.8.0/leaflet.zip && unzip leaflet.zip -d leaflet  && rm leaflet.zip && \ 
    wget -c https://github.com/CliffCloud/Leaflet.EasyButton/archive/refs/tags/v2.4.0.zip && unzip v2.4.0.zip && rm v2.4.0.zip && \ 
    cp Leaflet.EasyButton-2.4.0/src/* leaflet/ && rm -r Leaflet.EasyButton-2.4.0
EXPOSE 8080
CMD ["/bin/sh", "/OfflineMapExplorer/scripts/run.sh"]