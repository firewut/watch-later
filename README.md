# Youtube Watch Later Service

https://watchlater.chib.me runs this code to move videos from your **watch later** to a **playlist named watch later**.

## Docker

Docker run command will be:

```
docker run -d \
-e DEFORMIO_PROJECT= <https://your_watchlater.deform.io> \
-e DEFORMIO_TOKEN= <your_deformio_access_token> \
-e HOSTNAME= <watchlater.example.com> \
-e HTTP_DOMAIN=https://watchlater.example.com \
-e OAUTH2_CLIENT_ID=<OAUTH2_CLIENT_ID> \
-e OAUTH2_CLIENT_SECRET=<OAUTH2_CLIENT_SECRET> \
-e OAUTH2_REDIRECT_URI=<https://watchlater.example.com/oauth2callback> \
-e OAUTH2_AUTH_URI=https://accounts.google.com/o/oauth2/auth \
-e OAUTH2_TOKEN_URI=https://accounts.google.com/o/oauth2/token \
-e TEMPLATES_DIR=/var/templates/ \
-e HASH_KEY=zQJXqb6FvgVCMkEuapxrZDtcyeUmiLYG7P \
-e BLOCK_KEY=TxKfiZhQHUkevwEjM2PaXcY9zubVFsop \
-p 80:9001 \
firewut/watchlater
```

All you need is to replace some variables.

  * `HOSTNAME` - for example: `watchlater.example.com`
  * `HTTP_DOMAIN` - for example: `https://watchlater.example.com`

Create or use existing [google oauth2 credentials](https://console.developers.google.com/apis/credentials?project=watch-later-1152):

  * `OAUTH2_CLIENT_ID`
  * `OAUTH2_CLIENT_SECRET`
  * `OAUTH2_REDIRECT_URI`

You can also regenerate random values:

  * `HASH_KEY` - 34 letters
  * `BLOCK_KEY` - 32 letters


Last step is to register in [https://deform.io](https://deform.io):

  * `DEFORMIO_PROJECT` - project name. For example: `watchlater.deform.io`
  * `DEFORMIO_TOKEN` - token's `id`. You will need a token which has a permissions to write. Your `First project Token` will fit.
