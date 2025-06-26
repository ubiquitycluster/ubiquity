# Ingress Steering and sticky sessions
We implement the following ingress steering and sticky sessions in Ubiquity:

```
nginx.ingress.kubernetes.io/affinity: "cookie"
nginx.ingress.kubernetes.io/session-cookie-name: "route"
nginx.ingress.kubernetes.io/session-cookie-expires: "172800"
nginx.ingress.kubernetes.io/session-cookie-max-age: "172800"
```
The reason for this is that we use a high-availability ingress instance spread over multiple nodes. This means that the ingress instance can be on a different node for each request. This can cause problems with the session. The above configuration ensures that the session is always on the same node.

# Here is the explanation for the code above:
1. Annotations to set up the sticky session are described in the official documentation.
2. The sticky session is implemented using the cookie method.
3. The cookie name is set to route.
4. The cookie expires after 172800 seconds, which is 48 hours.
5. The cookie max age is set to 172800 seconds, which is 48 hours.
6. The sticky session is enabled for the ingress defined.

# To test the sticky session using the curl command. Here's the command that we're going to use:
`curl -v -b cookie.txt -c cookie.txt http://<INGRESS_IP>/hello`
        
The command above will send a GET request to the Ingress resource. Let's break down the command above:
- The `-v` flag is used to enable the verbose mode.
- The `-b` flag is used to read cookies from the file.
- The `-c` flag is used to write cookies to the file.
- The `http://<INGRESS_IP>/hello` part is the URL that we want to send the request to. Replace <INGRESS_IP> with the IP address of your Ingress resource.

Let's test the sticky session using the curl command:
`curl -v -b cookie.txt -c cookie.txt http://<INGRESS_IP>/hello`

Here's an example output:

```
* Added cookie route="route1" for domain <INGRESS_IP>, path /, expire 1641010396
* Added cookie route="route1" for domain <INGRESS_IP>, path /, expire 1641010396
*   Trying <INGRESS_IP>...
* TCP_NODELAY set
* Connected to <INGRESS_IP> (<INGRESS_IP>) port 80
```