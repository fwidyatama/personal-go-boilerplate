#!/bin/sh

CONFIG=config.yaml
ENVFILE=.env

echo "# Generated from $CONFIG" > $ENVFILE

# App
echo "APP_NAME=$(yq '.app.name' $CONFIG)" >> $ENVFILE
echo "APP_VERSION=$(yq '.app.version' $CONFIG)" >> $ENVFILE
echo "APP_PORT=$(yq '.app.port' $CONFIG)" >> $ENVFILE
echo "APP_ENV=$(yq '.app.env' $CONFIG)" >> $ENVFILE

# Log
echo "LOG_LEVEL=$(yq '.log.level' $CONFIG)" >> $ENVFILE
echo "LOG_FORMAT=$(yq '.log.format' $CONFIG)" >> $ENVFILE

# Database
echo "DB_HOST=$(yq '.database.host' $CONFIG)" >> $ENVFILE
echo "DB_PORT=$(yq '.database.port' $CONFIG)" >> $ENVFILE
echo "DB_NAME=$(yq '.database.name' $CONFIG)" >> $ENVFILE
echo "DB_USER=$(yq '.database.user' $CONFIG)" >> $ENVFILE
echo "DB_PASSWORD=$(yq '.database.password' $CONFIG)" >> $ENVFILE
echo "DB_SSL_MODE=$(yq '.database.ssl_mode' $CONFIG)" >> $ENVFILE
echo "DB_MAX_OPEN_CONNS=$(yq '.database.max_open_conns' $CONFIG)" >> $ENVFILE
echo "DB_MAX_IDLE_CONNS=$(yq '.database.max_idle_conns' $CONFIG)" >> $ENVFILE

# Redis
echo "REDIS_HOST=$(yq '.redis.host' $CONFIG)" >> $ENVFILE
echo "REDIS_PORT=$(yq '.redis.port' $CONFIG)" >> $ENVFILE
echo "REDIS_PASSWORD=$(yq '.redis.password' $CONFIG)" >> $ENVFILE
echo "REDIS_DB=$(yq '.redis.db' $CONFIG)" >> $ENVFILE

# Kafka
echo "KAFKA_BROKERS=$(yq e '.kafka.brokers | join(",")' $CONFIG)" >> $ENVFILE
echo "KAFKA_TOPIC_PREFIX=$(yq '.kafka.topic_prefix' $CONFIG)" >> $ENVFILE
echo "KAFKA_GROUP_ID=$(yq '.kafka.group_id' $CONFIG)" >> $ENVFILE
echo "KAFKA_AUTO_OFFSET=$(yq '.kafka.auto_offset' $CONFIG)" >> $ENVFILE
echo "KAFKA_MAX_RETRIES=$(yq '.kafka.max_retries' $CONFIG)" >> $ENVFILE
echo "KAFKA_RETRY_BACKOFF=$(yq '.kafka.retry_backoff' $CONFIG)" >> $ENVFILE

# Auth
echo "JWT_SECRET=$(yq '.auth.jwt_secret' $CONFIG)" >> $ENVFILE
echo "REFRESH_SECRET=$(yq '.auth.refresh_secret' $CONFIG)" >> $ENVFILE
echo "ACCESS_TTL=$(yq '.auth.access_ttl' $CONFIG)" >> $ENVFILE
echo "REFRESH_TTL=$(yq '.auth.refresh_ttl' $CONFIG)" >> $ENVFILE

echo "Generated $ENVFILE from $CONFIG"