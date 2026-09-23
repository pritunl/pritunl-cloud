#!/bin/bash
set -e

rsync --human-readable --archive --xattrs --progress --delete --exclude "/node_modules/*" --exclude "/jspm_packages/*" --exclude "app/*.js" --exclude "app/*.js.map" --exclude "app/**/*.js" --exclude "app/**/*.js.map" /home/cloud/git/pritunl-cloud/www/ $NPM_SERVER:/home/cloud/pritunl-cloud-www/

ssh cloud@$NPM_SERVER "
cd /home/cloud/pritunl-cloud-www/
rm -rf node_modules
npm install
cd ./node_modules/@github/webauthn-json/dist/
ln -sf ./esm/* ./
cd ../../../../
"

scp $NPM_SERVER:/home/cloud/pritunl-cloud-www/package.json /home/cloud/git/pritunl-cloud/www/package.json
scp $NPM_SERVER:/home/cloud/pritunl-cloud-www/package-lock.json /home/cloud/git/pritunl-cloud/www/package-lock.json
rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-cloud-www/node_modules/ /home/cloud/git/pritunl-cloud/www/node_modules/
rsync --human-readable --archive --xattrs --progress --delete --exclude "/node_modules/*" --exclude "/jspm_packages/*" --exclude "app/*.js" --exclude "app/*.js.map" --exclude "app/**/*.js" --exclude "app/**/*.js.map" /home/cloud/git/pritunl-cloud/www/ $NPM_SERVER:/home/cloud/pritunl-cloud-www/

ssh cloud@$NPM_SERVER "
cd /home/cloud/pritunl-cloud-www/
sh build.sh
"

rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-cloud-www/dist/ /home/cloud/git/pritunl-cloud/www/dist/
rsync --human-readable --archive --xattrs --progress --delete $NPM_SERVER:/home/cloud/pritunl-cloud-www/dist-dev/ /home/cloud/git/pritunl-cloud/www/dist-dev/
