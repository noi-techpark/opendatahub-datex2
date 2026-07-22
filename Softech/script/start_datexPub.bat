docker stop datexPub
docker container prune --force

docker run -d --name datexPub --restart on-failure --pull always -e TZ=Europe/Rome -v "E:\TEST\NODO_DATEX\datexPub.appsettings.json":/app/appsettings.json -v "E:\TEST\NODO_DATEX\datexPub.NLog.config":/app/NLog.config -v "E:\TEST\NODO_DATEX\log":/app/log -v "E:\TEST\NODO_DATEX\pubblicazioni":/app/pubblicazioni 192.168.1.245:5000/datexpub:latest
pause
