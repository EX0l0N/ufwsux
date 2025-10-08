# ufwsux

## Short for: Your Firewall shows your experience\!

(Not what you think\!)

### So what is this about?

Nowadays a lot of inexperienced hobby admins configure guest WiFi for bars and so on.  
Most of them seem to have a feeling that they need to configure some “extra security”, because of providing a publicly accessible WiFi.  
So how secure can you get, when you’re unable to limit your user base in any way?  
Well, you could at least control what they do.  
You don’t want your customers to sit at your place playing games, ignoring everyone else after all. So let’s limit the connection to surfing and mailing. That should be enough for anyone, even for those who casually want to work remotely, right?

## Well, no.

Most of us remoters rely on tools using other ports than `80`, `443`, `25` (and whatever ports may be in use for sending and receiving mail)\!  
With GitHubs move to disable https access to repositories completely every remote IT worker needs to have at least access to port `22` for `ssh`.  
Guess what. That’s neither “*surfing*” or “*mailing*”.

## ufwsux solves that problem by piping a tcp connection through plain http.

But it doesn’t do that the simple way, occupying the port and by doing so preventing any webserver to run there.  
It’s routable by a webserver like nginx, so you just need to reverse proxy a single URL, coexisting with whatever you want to host there.

Here's my own `nginx` location block for reference:

```
        location /ufwsux {
                proxy_pass http://localhost:1199/; // <- doesn't need to be localhost

                proxy_http_version 1.1; // <- http 1.1 required!

                proxy_set_header Upgrade $http_upgrade;
                proxy_set_header Connection $http_connection;

                proxy_buffering off;

                proxy_read_timeout 3600s;
                proxy_send_timeout 3600s;
        }
```

This is meant to go to unencrpted plain http servers. Don't do multiple layers of encryption if you don't have to.  
Don't forget to have `ufwsuxd` run in a `tmux` or, screen `terminal`, write a `systemd` unit if that's your kind of kink.
That's all there is to do.

The client side part acts like an `ssh` proxy command just like `nc`, or `netcat` is often used.

With a `ufwsux` webserver at hand you just pass the webserver url to the client command and `ssh` will pass the destination host and port.

Equipped like that you can now connect `ssh` through any http supporting WiFi and with that also `ssh` tunnels, even layer2 and layer3, meaning virtual ethernet interfaces, or routing interfaces.  
Yes, `ssh` can do that. Read it up.

An example `ssh` invocation looks like this:

```
ssh -o ProxyCommand="./ssh-proxy <ufwsuxd server> %h %p" <user>@<targethost>
```

## How does it work?

`ufwsuxd` (the server part) listens at port 1199 and waits for incoming http connections from the webserver. It will take all the required information from header fields of that request which also serves as a connection upgrade request.  
The server will then drop to plain tcp instead of doing something fancy like establishing a websocket.  
On the other side the server will open a tcp connection to the requested host and port.  
Once that is done it will act as a pipe between the two connections effectively tunnelling your connection through http.

# Is it secure?

Well, sort of. The connection is secured by a shared secret which you need to define before compiling. Client and server will try to match a hash that is based on connection parameters, the shared secret and a timestamp, so replay attacks only make sense for a short period, before the header becomes invalid to use.

You need to create the file `tokens/secret.go` and make it look like this (don't use my example secret here!):

```
package tokens

// Secret used for HMAC token generation
const secret = "your-very-secure-and-long-secret"
```

The connection itself is **not** encrypted.

You are expected to use it for `ssh` or some kind of **VPN** if you like.  
Having `ufwsux` encrypt for that usecase, would be an encrypted connection in an encrypted connection, which doesn’t add any security, but diminishes performance.

If you use it for plain tcp, your data will be readable in plaintext. This would be your fault.

## Precompiled binaries?

As of now, with the shared secret being compiled in, providing precompiled binaries would also distribute a common shared secret, which would be pointless.

I want to keep this tool simple and avoid any config files, complex parameters or key management systems in the pipeline, so you need to compile it yourself after setting the secret.

If you're interested in tools like this, running `go build` is probably way below your skill level anyway.  
And by the way, `ufwsux` has no dependencies, which would need to be installed before compiling.

I may write a `Makefile` at some point later.

# Happy tunnelling!
