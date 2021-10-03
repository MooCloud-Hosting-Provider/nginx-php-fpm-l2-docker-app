ARG ARG_DIRPATH_SCRIPTS_PKG=/app/pkg/scripts

FROM ubuntu:20.04
ENV DEBIAN_FRONTEND=noninteractive

ARG ARG_DIRPATH_SCRIPTS_PKG

RUN apt-get update
RUN apt-get upgrade -y

RUN apt-get install -y bash
RUN apt-get install -y gcc
RUN apt-get install -y shc
RUN apt-get install -y g++
RUN apt-get install -y libc6-dev
RUN apt-get install -y wget

RUN wget https://golang.org/dl/go1.17.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.17.linux-amd64.tar.gz

RUN mkdir -p /app/pkg
RUN mkdir -p ${ARG_DIRPATH_SCRIPTS_PKG}

RUN mkdir -p /app/pkg/phpParser
RUN mkdir -p /app/pkg/nginxParser

COPY moocli/*.go /app/pkg/
COPY moocli/phpParser/*.go /app/pkg/phpParser
COPY moocli/nginxParser/*.go /app/pkg/nginxParser

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=off ./usr/local/go/bin/go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o ${ARG_DIRPATH_SCRIPTS_PKG}/moocli /app/pkg/moocli.go

COPY scripts/start.sh /app/pkg
COPY cloudron/setup.sh /app/pkg

RUN shc -v -r -f /app/pkg/start.sh -o ${ARG_DIRPATH_SCRIPTS_PKG}/start
RUN shc -v -r -f /app/pkg/setup.sh -o ${ARG_DIRPATH_SCRIPTS_PKG}/setup

# restart docker image building

FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

ENV PHP_VERSION 7.4
ENV PHP_INI /etc/php/${PHP_VERSION}/cli/php.ini
ENV PHP_FPM_INI /etc/php/${PHP_VERSION}/fpm/php.ini

RUN apt-get update
RUN apt-get upgrade -y

RUN apt-get install -y ca-certificates
RUN apt-get install -y supervisor

RUN apt-get install -y nginx

RUN apt-get install -y php${PHP_VERSION}-dev
RUN apt-get install -y php${PHP_VERSION}-fpm
RUN apt-get install -y php${PHP_VERSION}-common
RUN apt-get install -y php${PHP_VERSION}-mbstring
RUN apt-get install -y php${PHP_VERSION}-xmlrpc
RUN apt-get install -y php${PHP_VERSION}-soap
RUN apt-get install -y php${PHP_VERSION}-gd
RUN apt-get install -y php${PHP_VERSION}-xml
RUN apt-get install -y php${PHP_VERSION}-intl
RUN apt-get install -y php${PHP_VERSION}-mysql
RUN apt-get install -y php${PHP_VERSION}-cli
RUN apt-get install -y php${PHP_VERSION}-ldap
RUN apt-get install -y php${PHP_VERSION}-zip
RUN apt-get install -y php${PHP_VERSION}-curl
RUN apt-get install -y php${PHP_VERSION}-imagick

ENV DEBIAN_FRONTEND=dialog

# important paths

ARG ARG_DIRPATH_SCRIPTS_PKG

ENV DIRPATH_SCRIPTS_PKG ${ARG_DIRPATH_SCRIPTS_PKG}

ENV DIRPATH_STATUS /app/data/status

ENV DIRPATH_NGINX /run/nginx
ENV DIRPATH_NGINX_PKG /app/pkg/nginx

ENV DIRPATH_NGINX_LOCATIONS /app/data
ENV DIRPATH_NGINX_LOCATIONS_PKG /app/pkg/data

ENV DIRPATH_PHP /run/php
ENV DIRPATH_PHP_PKG /app/pkg/php

ENV DIRPATH_SUPERVISORD /run/supervisord
ENV DIRPATH_SUPERVISORD_PKG /app/pkg/supervisord

ENV DIRPATH_MOOCLI_LICENSE /app/data/moocli
ENV FILEPATH_MOOCLI_LICENSE /app/data/moocli/license

ENV DIRPATH_FASTCGI_CACHE_DRIVE /app/data/cache
ENV DIRPATH_FASTCGI_CACHE_RAMDISK /media/ramdisk

ENV DIRPATH_WWW /app/data/www
ENV DIRPATH_WWW_PKG /app/pkg/www

ENV DIRPATH_CLOUDRON /app/data
ENV DIRPATH_CLOUDRON_PKG /app/pkg/data

RUN mkdir -p ${DIRPATH_PHP_PKG}
RUN mkdir -p ${DIRPATH_WWW_PKG}
RUN mkdir -p ${DIRPATH_NGINX_PKG}
RUN mkdir -p ${DIRPATH_SCRIPTS_PKG}
RUN mkdir -p ${DIRPATH_CLOUDRON_PKG}
RUN mkdir -p ${DIRPATH_SUPERVISORD_PKG}
RUN mkdir -p ${DIRPATH_NGINX_LOCATIONS_PKG}

# import the compiled binaries into the new docker image
COPY --from=0 ${DIRPATH_SCRIPTS_PKG}/* ${DIRPATH_SCRIPTS_PKG}

COPY nginx/nginx.conf ${DIRPATH_NGINX_PKG}
COPY nginx/proxy.conf ${DIRPATH_NGINX_PKG}
COPY nginx/mime.types ${DIRPATH_NGINX_PKG}
COPY nginx/default.conf ${DIRPATH_NGINX_PKG}
COPY nginx/fastcgi.conf ${DIRPATH_NGINX_PKG}
COPY nginx/fastcgi_params.conf ${DIRPATH_NGINX_PKG}
COPY nginx/custom/nginx_locations.conf ${DIRPATH_NGINX_LOCATIONS_PKG}

COPY php-fpm${PHP_VERSION}/php-fpm.conf ${DIRPATH_PHP_PKG}
COPY php-fpm${PHP_VERSION}/www.conf ${DIRPATH_PHP_PKG}
COPY php-fpm${PHP_VERSION}/custom/custom.conf ${DIRPATH_PHP_PKG}

COPY cloudron/variables.txt ${DIRPATH_CLOUDRON_PKG}

COPY supervisord/supervisord.conf ${DIRPATH_SUPERVISORD_PKG}/supervisord.conf

RUN echo "session.save_path = ${DIRPATH_PHP}/session" >> ${PHP_INI}
RUN echo "session.save_path = ${DIRPATH_PHP}/session" >> ${PHP_FPM_INI}
RUN sed --in-place "s|session.cookie_path = /|session.cookie_path = ${DIRPATH_PHP}/cookie|" ${PHP_INI}
RUN sed --in-place "s|session.cookie_path = /|session.cookie_path = ${DIRPATH_PHP}/cookie|" ${PHP_FPM_INI}

# initial test website to check correctness
# and have an example to know where to start
COPY test/www/* ${DIRPATH_WWW_PKG}

# outside port
EXPOSE 80

ENTRYPOINT .${DIRPATH_SCRIPTS_PKG}/start