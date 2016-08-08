FROM gliderlabs/alpine:3.4

USER root

RUN apk add --update openssl ca-certificates

ADD ./bin/project /bin/project
COPY src/project/templates /var/templates

EXPOSE 9001

VOLUME ["/var/log/project/"]

ENTRYPOINT /bin/project \
    --mode=$MODE \
	--http-host=0.0.0.0:9001 \
	--http-domain=$HTTP_DOMAIN \
	--access-log=$ACCESS_LOG \
	--info-log=$INFO_LOG \
	--error-log=$ERROR_LOG \
	--hostname=$HOSTNAME \
	--oauth2-client-id=$OAUTH2_CLIENT_ID \
	--oauth2-client-secret=$OAUTH2_CLIENT_SECRET \
	--oauth2-redirect-uri=$OAUTH2_REDIRECT_URI \
	--oauth2-auth-uri=$OAUTH2_AUTH_URI \
	--oauth2-token-uri=$OAUTH2_TOKEN_URI \
	--deformio-project=$DEFORMIO_PROJECT \
	--deformio-token=$DEFORMIO_TOKEN \
	--templates-dir=$TEMPLATES_DIR \
	--hash-key=$HASH_KEY \
	--block-key=$BLOCK_KEY \
	--sentry-dsn=$SENTRY_DSN