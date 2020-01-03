git pull origin master
docker stop tamer
docker rm tamer
docker build -t tamer ./
docker run -p 5000:5000 --name tamer -t tamer &
cd ..
./tech-db-forum perf -i "" -v 1.0