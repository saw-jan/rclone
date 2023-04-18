# S3 Proxy

## install, configure, run rclone
- clone the version of  `rclone` with the fix: `git clone --branch=forwardS3AccessKeyToWebDAV --single-branch https://github.com/JankariTech/rclone.git`
- build rclone. For me a simple `make` did the job, but here are more details on how to build it: https://rclone.org/install/#source
- start OwnCloud or Nextcloud as a WebDAV server
- configure the WebDAV server as a storage backend to rclone:
  - run `./rclone config`
  - select `n` for `New remote`
  - enter a name for the remote, e.g `OC`
  - select `45` (WebDAV) as Storage
  - give your WebDAV URL, in case of OC/NC its `<protocol><host>/remote.php/webdav/` e.g. http://localhost:8080/remote.php/webdav/
  - select OwnCloud or Nextcloud as vendor
  - leave the username blank
  - select `N`  to leave the password blank
  - skip over bearer_token question
  - skip over advance config question
  - accept the configuration
  - exit the config tool
- check `./rclone listremotes --long`
  it should show the name of the config e.g. `OC: webdav` 
- serve the data through the S3 protocol:  `./rclone serve s3 <backend name>:` e.g. `./rclone serve s3 OC:` (column after the name is needed)
  by default the server will listen to port `8080`. To listen on a different port, add `--addr localhost:8081`

### get an oauth token
- install the oauth2 app in ownCloud/Nextcloud
- in ownCloud/Nextcloud create a new oauth2 client with the redirect URL `http://localhost:9876/`
- create a json file called `oauth.json` and this content
```
   {
  "installed": {
    "client_id": "<client-id-copied-from-oauth2-app>",
    "project_id": "focus-surfer-382910",
    "auth_uri": "<owncloud-server-root>/index.php/apps/oauth2/authorize",
    "token_uri": "<owncloud-server-root>/index.php/apps/oauth2/api/v1/token",
    "client_secret": "<client-secret-copied-from-oauth2-app>",
    "redirect_uris": [
      "http://localhost:9876"
    ]
  }
}
```
- download and install `oauth2l` https://github.com/google/oauth2l#pre-compiled-binaries
- get an oauth2 token `./oauth2l fetch --credentials ./oauth.json --scope all --refresh --output_format bare`

### access WebDAV using S3:
every root folder becomes a bucket, every file below it is a key
 
### curl
- list all buckets (root folders): `curl --location 'http://localhost:8080/' -H'Authorization: AWS4-HMAC-SHA256 Credential=<oauth-access-token>/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;range;x-amz-date,Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024'`
  replace `<oauth-access-token>` with your token, that you got with `oauth2l` , the rest of the header has to exist, but the content doesn't matter currently
- list keys in a bucket (any file below the root folder): `curl --location 'http://localhost:8080/<bucket>?list-type=2' -H'Authorization: AWS4-HMAC-SHA256 Credential=<oauth-access-token>/20130524/us-east-1/s3/aws4_request,SignedHeaders=host;range;x-amz-date,Signature=fe5f80f77d5fa3beca038a248ff027d0445342fe2855ddc963176630326f1024'` 
 
 ### minio client
- install `mc`: https://min.io/docs/minio/linux/reference/minio-mc.html
- add the proxy as alias: `mc alias set mys3proxy http://localhost:8080 <oauth-access-token> anysecretkeyitdoesnotmatter`
  replace `<oauth-access-token>` with your token, that you got with `oauth2l`
- list buckets: `mc ls mys3proxy`
- list items in a bucket `mc ls mys3proxy/folder`
- upload a file: `mc cp <localfile> mys3proxy/<bucket (root folder)>/<keyname>`
- find files `mc find proxy/<bucket>/ --name "*.txt"`
- download all content of a bucket: `mc cp --recursive mys3proxy/<bucket> /tmp/dst/` (items that have a space in their name seem to have an issue)
