#!/bin/bash

if [[ ! -f ${DIRPATH_STATUS}/.initialized ]]; then
    echo "Fresh installation, setting up public folder..."
    
    mkdir -p ${DIRPATH_WWW}
    mkdir -p ${DIRPATH_STATUS}
    mkdir -p ${DIRPATH_CLOUDRON}
    mkdir -p ${DIRPATH_MOOCLI_LICENSE}
    mkdir -p ${DIRPATH_NGINX_LOCATIONS}
    mkdir -p ${DIRPATH_FASTCGI_CACHE_DRIVE}
    
    # test data, only at first install
    cp ${DIRPATH_WWW_PKG}/* ${DIRPATH_WWW}

    touch ${DIRPATH_STATUS}/.initialized
    echo "Initial setup complete."
fi

mkdir -p ${DIRPATH_PHP}
mkdir -p ${DIRPATH_NGINX}
mkdir -p ${DIRPATH_PHP}/cookie
mkdir -p ${DIRPATH_PHP}/session
mkdir -p ${DIRPATH_SUPERVISORD}

cp ${DIRPATH_NGINX_PKG}/*.conf ${DIRPATH_NGINX}
cp ${DIRPATH_NGINX_PKG}/mime.types ${DIRPATH_NGINX}

cp ${DIRPATH_PHP_PKG}/*.conf ${DIRPATH_PHP}
cp ${DIRPATH_SUPERVISORD_PKG}/*.conf ${DIRPATH_SUPERVISORD}
cp ${DIRPATH_CLOUDRON_PKG}/variables.txt ${DIRPATH_CLOUDRON}
cp ${DIRPATH_NGINX_LOCATIONS_PKG}/nginx_locations.conf ${DIRPATH_NGINX_LOCATIONS}

chmod a+x ${DIRPATH_SCRIPTS_PKG}/setup
${DIRPATH_SCRIPTS_PKG}/setup

exec supervisord -c ${DIRPATH_SUPERVISORD}/supervisord.conf