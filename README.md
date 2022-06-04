# OfflineMapExplorer
![](/preview.png)

## О проекте
&nbsp;&nbsp;&nbsp;&nbsp;Данный проект использует библиотеку [Mapnik 3.x.x](https://mapnik.org/) для рендеринга, а для отрисовки изображений у клиента используется библиотека leaflet. Официальной обёртки для Go библиотека Mapnik не имеет. Однако сущетвует [C API](https://github.com/springmeyer/mapnik-c-api) для Mapnik , написанный одним из разработчиков Mapnik. На основе этого API существует [врапер для Go](https://github.com/omniscale/go-mapnik). Однако данный репозиторий давно не обновлялся, а потому возникла проблема с загрузкой шрифтов для Mapnik (при загрузке не учитываются поддиректории), а потому был сделан [форк с исправлением данной проблемы](https://github.com/KanesRF/go-mapnik).
## Описание компонентов
![](/сomponents.png)
&nbsp;&nbsp;&nbsp;&nbsp;Упрощённая схема взаимодействия компонентов представлена на рисунке. Карты чаще всего можно найти в формате .osm, [на одном из ресурсов](http://download.geofabrik.de/) имеются карты как для всей планеты, так и для отдельных стран, регионов и т.д. Данные из .osm файлов экспортируются в базу данных Postgresql с плагином Postgis с помощью утилиты osm2pgsql. Однако этого будет недостаточно для рендерига карт.

&nbsp;&nbsp;&nbsp;&nbsp;Библиотека Mapnik принимает на вход XML файл с описанием слоёв, шрифтов, и прочих конфигураций. Таким образом, в зависимости от XML файла может меняться стиль отрисованной карты. В данном проекте был использован [стиль OpenStreetMap](https://github.com/gravitystorm/openstreetmap-carto). Для генерации XML файла используется средство CartoCSS. Помимо XML файла генерируются данные, используемые при отрисовке (ссылки на них имеются в XML файле). [Проект со стилем OpenStreetMap](https://github.com/gravitystorm/openstreetmap-carto) также содержит скрипты, которые преобразуют данные при экспорте карты в Postgis, а также скрипты для создания индексов после импорта.

#### _cmd/prerender_
&nbsp;&nbsp;&nbsp;&nbsp; Рендеринг клеток по запросу от пользователя может быть очень ресурсоёмким и занимать много времени, а потому при большом приближения, или клетки с большим  числом объектов имеет смысл отрендерить заранее. Для этого может быть использована программа _cmd/prerender_, которая рендерит клетки в указанном "радиусе" от заданной точки для указанного приближения. Отрисованные клетки сохраняются в prerendered/_z_/. Аргументы, которые можно передать в _cmd/prerender_:
- z - уровень приближения
- x - коррдината X
- y - координата Y
- r - радиус отрисовки клеток вокруг точки (X, Y)
- f - путь к XML файлу для Mapnik

Пример вызова
```sh
./cmd/prerender -z 5 -x 10 -y 5 -r 5 -f style.xml
```

#### _cmd/server_
&nbsp;&nbsp;&nbsp;&nbsp; Сервер принимиет запросы от клиента в формате _https://localhost:8080/z/x/y.png_. Часть клеток заранее отрисована и находится в директории _prerendered_. Те клетки, что не были заранее отрисованы, будут рендериться "на лету". Для избежания повторного рендеринга часто запрашиваемых клеток был использован [LRU кэш](https://github.com/hashicorp/golang-lru) . При посещении ссылки _https://localhost:8080/_ будет передан html файл. Сервер принимает следующие аргументы:

- z - уровень приближения, при котором нужно искать заранее отрисованные клетки
- pool - число объектов Map библиотеки Mapnik, которые отрисовывают клетки
- max_zoom - максимально доступный уровень приближения
- f - путь к XML файлу для Mapnik

Пример вызова
```sh
./cmd/server -f style.xml -pool 4 -max_zoom 10 -z 6
```

## Структуры и алгоритм работы програм
&nbsp;&nbsp;&nbsp;&nbsp; Кодовая база данного проекта довольно мала. За отрисовку отвечает пакет _render_, который реализован в файле render/render.go. Данный пакет содержит следующие структуры:

- Структура Coords, которая описывает клетку на карте
``` go
type Coords struct {
	X, Y, Z int
}
```
-   Структура TileRender. Данная структура содержит указатель на объект Map библиотеки Mapnik и указатель на параметры для рендеринга (в данном проекте используются только для указания формата генерируемого изображения). Структура mapnik.Map имеет методы Render и RenderToFile для отрисовки и Load для загрузки XML. 
``` go
type TileRender struct {
	m    *mapnik.Map
	opts *mapnik.RenderOpts
}
```
- Структура Queue, которая содержит слайс указателей на структуру TileRender (который используется как стэк свободных для использования объектов) и условную переменную. В ходе работы горутины при запросе берут свободный объект TileRender, реднерят с его помощью клетку, после чего кладут обратно в стэк.
``` go
type Queue struct {
	queue       []*TileRender
	isAvailable sync.Cond
}
```
&nbsp;&nbsp;&nbsp;&nbsp; Исходный код програмы для пререндеринга состоиз из файлов utils/tile_prerender.go и render/render.go. Алгоритм работы данной программы предельно прост:
1. Парсинг аргументов.
2. Инициализация структуры Queue путём паралельного создания горутинами структур TileRender и их инициализации.
3. Создание пула горутин для отрисовки. Горутины при получении координат получают свободный объект TileRender, рендерят клетки и освобождают объект TileRender.
4. Поочерёдная генерация файлов-клеток путём передачи пулу горутин координат для отрисовки.
5. Закрытие канала для отправки координат пулу горутин и ожидание завершение рендеринга.

&nbsp;&nbsp;&nbsp;&nbsp; Исходный код сервера состоиз из файлов server/server.go и render/render.go. Алгоритм работы сервера:
1. Парсинг аргументов.
2. Инициализация структуры Queue путём паралельного создания горутинами структур TileRender и их инициализации.
3. Инициализация параметров https сервера и его запуск.

Алгоритм обработки запроса на получение клетки следующий:
1. Считать координаты.
2. Проверить корректность координат.
3. Если в кэше есть клетка с такой координатой, то отправить пользователю.
4. Если для данного уровня приближения предусмотрен пререндеринг, то считать клетку из файла, добавить в кэш и отправить пользователю.
5. Если клетка не была заранее отрисована и не находится в кэше, то:
 - Получить объект TileRender.
 - Отрисовать клетку.
 - Вернуть объект TileRender.
6. Добавить клетку в кэш и отправить пользователю. 

&nbsp;&nbsp;&nbsp;&nbsp; Пулл объектов TileRender необходим, т.к. объекты mapnik.Map одновременно могут отрисовывать только одну клетку. Паралельная инициализация этих объектов с множеством считывания XML файла связана с тем, что нет API для "глубокого" копирования этих объектов для Go (т.к. нет в C API). _По моим наблюдениям (и по обсуждениям на форумах) при отрисовке объект mapnik.Map может занимать 2-3.5 Gb оперативной памяти. Имеющийся API не даёт гибкости для оптимизации процесса отрисовки. Разработчики поддерживают обёртки для Python и Nodejs, а сама библиотека написана на C++, а потому зачастую проекты состоят из двух (и более) сервисов, один из которых написан на C++/Python/Nodejs и отвечает за отрисовку, например [плагин для apache с отдельным сервисом](https://github.com/openstreetmap/mod_tile)._

&nbsp;&nbsp;&nbsp;&nbsp; Для взаимодействия с сервером используется web-клиент, написанный с использованием библиотеки leaflet и плагина для неё [easy button](https://github.com/CliffCloud/Leaflet.EasyButton). Исходный код расположен в директории js. Клиент имеет три кнопки: уменьшить, увеличить приближение и вернуться к центру (в данном случае - город Москва) при текущем уровне приближения.

## Подготовка системы

Данная сборка тестировалась на ubuntu 22

Установка python, go, компиляторов
``` sh
sudo apt-get install -y golang g++ gcc clang python python3 python3-pip
```
Установка Mapnik
``` sh
sudo apt-get install -y git autoconf libtool libxml2-dev libbz2-dev \
libgeos-dev libgeos++-dev libproj-dev gdal-bin libgdal-dev g++ \
libmapnik-dev mapnik-utils python3-mapnik
apt-get install -y libharfbuzz-dev mapnik-utils
```
Загрузить стиль OpenStreetMap для дальнейшей генерации с помощью CartoCSS XML файла для Mapnik
``` sh
git clone https://github.com/gravitystorm/openstreetmap-carto.git
cd openstreetmap-carto
```
Установка шрифтов
``` sh
sudo apt-get install -y fonts-noto-cjk fonts-noto-hinted fonts-noto-unhinted fonts-hanazono fonts-unifont
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
```
Установка CartoCSS
```sh
sudo npm install -g carto
npm install mapnik-reference
node -e "console.log(require('mapnik-reference'))"
```
Генерация XML
```sh
carto -a "3.0.22" project.mml > style.xml
```
Установка Postgresql и Postgis
```sh
sudo apt-get install -y postgresql postgis postgresql-contrib postgresql-postgis postgresql-postgis-scripts
sudo service postgresql start
```
Подготовка Postgresql к загрузке геоданных
```sh
sudo -u postgres createuser -s $USER
sudo -u postgres createdb gis
psql -d gis -c 'CREATE EXTENSION postgis; CREATE EXTENSION hstore;'
psql -d gis -c 'ALTER TABLE geometry_columns OWNER TO postgres;'
psql -d gis -c 'ALTER TABLE spatial_ref_sys OWNER TO  postgres;'
```
Загрузка карт .osm. Этот проект тестировался на Центральном федеральном округе РФ
```sh
wget -c http://download.geofabrik.de/russia/central-fed-district-latest.osm.pbf
```
Установка osm2pqsql и загрузка геоданных из .osm в Postgis
```sh
sudo apt install -y osm2pgsql
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
```
Генерация необходимых файлов для отрисовки клеток
```sh
python3 -m pip install psycopg2-binary
scripts/get-external-data.py
```
Создание индексов в базе данных
```sh
HOSTNAME=localhost
psql -d gis -f indexes.sql
```
Теперь нужно скопировать директории _data, style, symbols, patterns_ и _style.xml_ в корень проекта.
Далее нужно скачать библиотеку [leaflet](https://leafletjs.com/download.html) и расположить в js/leaflet. После этого нужно скачать плагин [EasyButton](https://github.com/CliffCloud/Leaflet.EasyButton) и поместить содержимое src в js/leaflet.
Можно выполнить сборку
```sh
make all
```






