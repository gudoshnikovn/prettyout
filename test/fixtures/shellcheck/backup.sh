#!/bin/bash

BACKUP_DIR=/var/backups
SOURCE_DIR=/var/www

do_backup() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file=$BACKUP_DIR/backup_$timestamp.tar.gz
    cd $SOURCE_DIR
    tar czf $backup_file .
    echo "Backup saved to $backup_file"
}

do_backup
