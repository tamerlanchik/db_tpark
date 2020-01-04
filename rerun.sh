git pull origin master
docker stop tamer
docker rm tamer
docker build -t tamer ./
#docker run -p 5000:5000 --name tamer -t tamer &
docker run -d --memory 1G --log-opt max-size=1M --log-opt max-file=3 --name tamer -p 5000:5000 tamer
cd ..
./tech-db-forum func --wait=180
#./tech-db-forum fill --timeout=900
./tech-db-forum perf -i "" --duration=600 --step=60 -v 1.0
#./tech-db-forum perf -i "" -v 1.0
#Jan  4 11:27:20 testdbt kernel: Out of memory: Kill process 20715 (tech-db-forum) score 506 or sacrifice child
#Jan  4 11:27:20 testdbt kernel: Killed process 20715 (tech-db-forum), UID 0, total-vm:1587388kB, anon-rss:980428kB, file-rss:0kB, shmem-rss:0kB