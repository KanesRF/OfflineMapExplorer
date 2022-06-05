cd /OfflineMapExplorer
sudo service postgresql start
service postgresql start
export GOMAXPROCS=4
ps -e | grep postgres
./cmd/server -f style.xml -pool 4 -max_zoom 10 -z 7