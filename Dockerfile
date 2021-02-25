FROM golang:1.15.2-buster AS build

# build generators
WORKDIR /opt/build
# Copy code to docker-container
COPY . /opt/build

RUN go build ./cmd/main.go

FROM ubuntu:20.04 AS release

MAINTAINER Vlad Amelin

# Make the "en_US.UTF-8" locale so postgres will be utf-8 enabled by default
RUN apt-get -y update && apt-get install -y tzdata

ENV TZ=Russia/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone



# Install postgresql
RUN apt-get update -y && apt-get install -y postgresql postgresql-contrib

# Run the rest of the commands as the ``postgres`` user created by the ``postgres-$PGVER`` package when it was ``apt-get installed``
USER postgres

# Create a PostgreSQL role named ``docker`` with ``docker`` as the password and
# then create a database `docker` owned by the ``docker`` role.
RUN /etc/init.d/postgresql start &&\
    psql --command "ALTER USER postgres WITH PASSWORD 'docker';" &&\
    createdb -O postgres Timetable &&\
    /etc/init.d/postgresql stop

# Expose the PostgreSQL port
EXPOSE 5432

# Add VOLUMEs to allow backup of config, logs and databases
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

# Back to the root user
USER root

# Define server port
EXPOSE 5000

# Define built server
WORKDIR /usr/src/app

COPY ./configs configs
#COPY ./dbconfig.yml dbconfig.yml
COPY ./scripts/init.sql init.sql

COPY --from=build /opt/build/main .

ENV PGPASSWORD docker
CMD service postgresql start &&  psql  --host=localhost --dbname=Timetable --username=postgres --file=init.sql -p 5432 -a -q && ./main