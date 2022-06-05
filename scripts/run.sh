cd /OfflineMapExplorer
sudo service postgresql start
sleep 60
sudo service postgresql restart
export GOMAXPROCS=8
ps -e | grep postgres
./cmd/server -f style.xml -pool 8 -max_zoom 10 -z 7