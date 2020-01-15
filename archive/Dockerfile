FROM golang:1.13-stretch AS builder

WORKDIR /usr/src/app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -v

FROM ubuntu:19.04
ENV DEBIAN_FRONTEND=noninteractive
ENV PGVER 11
ENV PORT 5000
ENV POSTGRES_HOST localhost
ENV POSTGRES_PORT 5432
ENV POSTGRES_DB forum
ENV POSTGRES_USER forum
ENV POSTGRES_PASSWORD forum
EXPOSE $PORT

RUN apt-get update && apt-get install -y postgresql-$PGVER

USER postgres

COPY forum.sql .

RUN echo "host all  all    0.0.0.0/0  trust" >> /etc/postgresql/$PGVER/main/pg_hba.conf
RUN  echo 'local all forum trust' | cat - /etc/postgresql/$PGVER/main/pg_hba.conf > /etc/postgresql/$PGVER/main/pg_hba.conf.bak && mv /etc/postgresql/$PGVER/main/pg_hba.conf.bak /etc/postgresql/$PGVER/main/pg_hba.conf

RUN service postgresql start &&\
    psql --command "CREATE USER forum WITH SUPERUSER PASSWORD 'forum';" &&\
    createdb -O forum forum &&\
    psql -U forum forum < ./forum.sql &&\
    service postgresql stop

RUN echo "include_dir='conf.d'" >> /etc/postgresql/$PGVER/main/postgresql.conf
ADD postgres.conf /etc/postgresql/$PGVER/main/conf.d/basic.conf

EXPOSE 5432

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]


COPY --from=builder /usr/src/app/main .
CMD service postgresql start && ./main