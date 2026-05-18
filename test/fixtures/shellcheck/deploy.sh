DEPLOY_DIR=/var/www/app
DEPLOY_USER=deploy
BACKUP_CMD=`tar czf /tmp/backup.tar.gz $DEPLOY_DIR`

deploy() {
    echo "Deploying to $DEPLOY_DIR"
    cp -r ./dist $DEPLOY_DIR
    chown -R $DEPLOY_USER:$DEPLOY_USER $DEPLOY_DIR
}

deploy
