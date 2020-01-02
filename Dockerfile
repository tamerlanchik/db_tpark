FROM golang:1.11-stretch AS build

# Копируем исходный код в Docker-контейнер
#ADD ./ /go/src/db_tpark
COPY ./ /go/src/db_tpark
WORKDIR /go/src/db_tpark
RUN go get -v ./...

# Собираем генераторы
#WORKDIR /opt/build/db_tpark
#RUN go mod vendor
#RUN go install ./vendor/github.com/go-swagger/go-swagger/cmd/swagger
#RUN go install ./vendor/github.com/jteeuwen/go-bindata/go-bindata
#RUN go get ./...

# Собираем и устанавливаем пакет
#RUN go generate -x tools.go
#RUN go install ./main.go
RUN go build -o forum ./main.go && cp ./forum /go/bin/


FROM ubuntu:18.04 AS release

MAINTAINER Artem V. Navrotskiy

#
# Установка postgresql
#
ENV PGVER 10
RUN apt -y update && apt install -y postgresql-$PGVER

# Run the rest of the commands as the ``postgres`` user created by the ``postgres-$PGVER`` package when it was ``apt-get installed``
USER postgres

COPY --from=build go/src/db_tpark/init.sql /

RUN echo "host all  all    0.0.0.0/0  trust" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN  echo 'local all docker trust' | cat - /etc/postgresql/$PGVER/main/pg_hba.conf > /etc/postgresql/$PGVER/main/pg_hba.conf.bak && mv /etc/postgresql/$PGVER/main/pg_hba.conf.bak /etc/postgresql/$PGVER/main/pg_hba.conf

# Create a PostgreSQL role named ``docker`` with ``docker`` as the password and
# then create a database `docker` owned by the ``docker`` role.
RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER docker WITH SUPERUSER PASSWORD 'docker';" &&\
    createdb -O docker dbtpark &&\
    psql -U docker dbtpark < ./init.sql &&\
    /etc/init.d/postgresql stop

# Adjust PostgreSQL configuration so that remote connections to the
# database are possible.
#RUN echo "host all  all    0.0.0.0/0  trust" >> /etc/postgresql/$PGVER/main/pg_hba.conf
#RUN  echo 'local all docker trust' | cat - /etc/postgresql/$PGVER/main/pg_hba.conf > /etc/postgresql/$PGVER/main/pg_hba.conf.bak && mv /etc/postgresql/$PGVER/main/pg_hba.conf.bak /etc/postgresql/$PGVER/main/pg_hba.conf

# And add ``listen_addresses`` to ``/etc/postgresql/$PGVER/main/postgresql.conf``
RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "fsync = off" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "synchronous_commit = off" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "full_page_writes = off" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "full_page_writes = off" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "shared_buffers = 4000Mb" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "effective_cache_size = 10000Mb" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "work_mem = 10Mb" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\
    echo "wal_buffers = 1gMb" >> /etc/postgresql/$PGVER/main/postgresql.conf &&\



# Expose the PostgreSQL port
EXPOSE 5432

# Add VOLUMEs to allow backup of config, logs and databases
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

# Back to the root user
USER root

# Объявлем порт сервера
EXPOSE 5000

# Собранный ранее сервер
COPY --from=build go/bin/forum /usr/bin/
#COPY --from=build go/src/db_tpark/init.sql /
#
#RUN psql -U postgres dbtpark < ./init.sql


#COPY ./init.sql /

#
# Запускаем PostgreSQL и сервер
#
CMD service postgresql start && nohup /usr/bin/forum