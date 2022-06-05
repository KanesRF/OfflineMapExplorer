cd /OfflineMapExplorer
sudo service postgresql start
export GOMAXPROCS=4
until sudo pg_isready -p 5432 -U postgres
do
    sleep 30
    sudo service postgresql restart
    sleep 30
done
./cmd/server -f style.xml -pool 4 -max_zoom 10 -z 5
