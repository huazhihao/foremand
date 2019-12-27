# foremand

[![Build Status](https://travis-ci.org/huazhihao/foremand.svg?branch=master)](https://travis-ci.org/huazhihao/foremand)
[![GoDoc](https://godoc.org/github.com/huazhihao/foremand?status.svg)](https://godoc.org/github.com/huazhihao/foremand)

`foremand` = `foreman` + `etcd`

## Quick Examples

This short example assumes foremand, etcd and etcdctl are installed locally.

1. Start a `etcd` cluster in dev mode:

    ```shell
    $ etcd
    ```

1. Write data to the key in `etcd`:

    ```shell
    $ ETCDCTL_API=3 etcdctl put host1/app "python -m SimpleHTTPServer 8001"
    OK
    ```

1. Register to `etcd`:

    ```shell
    $ foremand -endpoints=http://127.0.0.1:2379 -prefix=host1
    INFO[0000] Initialing foremand                           endpoints="[http://127.0.0.1:2379]" prefix=host1
    INFO[0000] Starting foremand
    INFO[0000] forking                                       app=host1/app shell="python -m SimpleHTTPServer 8001"
    ```

1. Test `app` connectivity:

    ```shell
    $ curl http://127.0.0.1:8001
    <!DOCTYPE html PUBLIC "-//W3C//DTD HTML 3.2 Final//EN"><html>
    <title>Directory listing for /</title>
    <body>
    <h2>Directory listing for /</h2>
    ...
    </body>
    </html>
    ```
