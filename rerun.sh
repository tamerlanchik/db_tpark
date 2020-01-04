git pull origin master
docker stop tamer
docker rm tamer
docker build -t tamer ./
#docker run -p 5000:5000 --name tamer -t tamer &
docker run -d --memory 1G --log-opt max-size=1M --log-opt max-file=3 --name tamer -p 5000:5000 tamer
cd ..
./tech-db-forum func --wait=180
./tech-db-forum fill --timeout=900
./tech-db-forum perf -i "" --duration=600 --step=60 -v 1.0
#./tech-db-forum perf -i "" -v 1.0