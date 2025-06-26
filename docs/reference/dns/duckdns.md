# DuckDNS and Kubernetes
This guide pretends to explain a quick and easy way to get our applications in Kubernetes/AKS running exposed to internet with HTTPs using DuckDNS as domain name provider.

## Pre-requisites and versions:
- AKS cluster version: 1.21.7
- Helm 3
- Ingress-controller nginx chart version 4.0.16
- Ingress-controller nginx app version 1.1.1
- cert-manager version 1.2.0
- cert-manager DuckDNS webhook version 1.2.2

(1) Add ingress-controller Helm repo
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
(2) Update repository
helm repo update
(3) Install ingress-controller with Helm
With this command we will be installing the latest version from the repository.
helm install nginx-ingress ingress-nginx/ingress-nginx --namespace ingress --create-namespace 
(4) Verify the pods are running fine in our cluster
kubectl get pods -n ingress
You should see something like this:
NAME                                                      READY   STATUS    RESTARTS   AGE
nginx-ingress-ingress-nginx-controller-74fb55cbd5-hjvr9   1/1     Running   0          41m

(5) We need to verify our ingress-controller has a public IP assigned
kubectl get svc -n ingress
We should see something similar to this, the key part here is to have an IP assigned in "EXTERNAL-IP", this might take a few seconds to show, it is expected as in the background what is happening is that Azure is spinning a "Public IP" resource for you and assigning it to the AKS cluster.
NAME                                               TYPE           CLUSTER-IP     EXTERNAL-IP      PORT(S)                      AGE
nginx-ingress-ingress-nginx-controller             LoadBalancer   10.0.33.214    20.190.211.14   80:32321/TCP,443:30646/TCP   38m
(6) Deploy a test application
Now we will be deploying a testing application that will be running inside a pod with a service that we will use to access the pods. This might feel a bit overkill as we have only a single pod and having a service for a single pod seems a lot but keep in mind that pods can be rescheduled at any given moment and they can even change their IPs while a service doesnt, so reaching our pods using a service is the best (and the desired) option. This also scales better as if we have more pods we will still use the same service to reach them and the service will load balance between them.

This is the yaml file for our test application:
apiVersion: apps/v1
kind: Deployment
metadata:
  name: echo-app
  namespace: default
spec:
  selector:
    matchLabels:
      app: echo-app
  replicas: 2
  template:
    metadata:
      labels:
        app: echo-app
    spec:
      containers:
      - name: echo-app
        image: hashicorp/http-echo
        args:
        - "-text=Test 123!"
        ports:
        - containerPort: 5678
(7) Deploy our service
apiVersion: v1
kind: Service
metadata:
  name: echo-svc
  namespace: default
spec:
  ports:
  - port: 80
    targetPort: 5678
  selector:
    app: echo-app
(8) Let's deploy an ingress resource
Now we need to deploy an ingress resource, this will tell our ingress controller how to manage the traffic that will be arriving to the public IP of the ingress controller (the one from the step 5), basically we are telling it to forward the traffic from the "/" path to the service of our application.
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-echo
  namespace: ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
    nginx.ingress.kubernetes.io/use-regex: "true"
    nginx.ingress.kubernetes.io/rewrite-target: /$1
spec:
  rules:
  - http:
      paths:
      - path: /(.*)
        pathType: Prefix
        backend:
          service:
            name: echo-svc
            port:
              number: 80

We are telling the ingress controller to forward all the traffic of the port 80 to the service echo-svc on it's port 80.

(9) Let's test it all together
To test this we will be accessing the ingress using the public IP that we got in step 5:

Using a web browser, go to https://IP

Using the command line run curl https://IP

Adding certificates with cert-manager for duckDNS
So far all is good, the only (small :) ) detail is that our ingress has an IP and not a domain/subdomain which is a bit hard for humans to remember and our traffic is all going unencrypted over http, we don't have any security (yet).
We will be adding cert-manager to generate TLS certificates for us in our DuckDNS subdomain, cert-manager not only allows us to get certificates, it also rotate them when they are about to expire (and we can configure how ofter we want to expire/rotate them).

(10) Let's install cert-manager
helm install cert-manager jetstack/cert-manager --namespace cert-manager --version v1.2.0 --set 'extraArgs={--dns01-recursive-nameservers=8.8.8.8:53\,1.1.1.1:53}' --create-namespace --set installCRDs=true
After a moment it will be done creating the needed resources, we can verify this by checking the status of the pods in the cert-manager namespace:
kubectl get pods -n cert-manager

Something like the following should appear:
NAME                                            READY   STATUS    RESTARTS   AGE
cert-manager-6c9b44dd95-59b6n                   1/1     Running   0          47m
cert-manager-cainjector-74459fcc56-6dfn8        1/1     Running   0          47m
cert-manager-webhook-c45b7ff-hrcnx              1/1     Running   0          47m
(11) Do you need a domain for free? DuckDNS to the rescue!
With all this in place we are ready to request a TLS certificate for our site/application, but first we need to own a domain or a subdomain to point to our public IP (step 5) so we can reach our pods/service using a name instead of an IP.
Another very important point to note is that cert-manager will only provide certificates if we can proove we own the domain/subdomain (this is to avoid the possibility of anyone requesting a certificate for a well known domain like google.com), to do this it has two methods http-01 and dns-01, we will focus this time in dns-01 which basically works like this: cert-manager requests us to provide credentials to access the domain/subdomain (in DuckDNS this is a token), with that cert-manager will generate a random string and make a TXT record in the domain provider with the value of the random string generated, will wait a moment and will use public DNSs to query that TXT record, if cert-manager finds the TXT record with the correct value it means we own that domain/subdomain and then will remove the TXT record and generate a certificate for us for that domain/subdomain. This will end with a secret in our K8s/AKS cluster containing the certificate and the key for that domain/subdomain, that secret is the one we will tell ingress-controller to use to validate the https traffic reaching our ingress.

(11) Configuring our DuckDNS account
We need to go to https://www.duckdns.org/ and log in with our account/credentials (you have multiple alternatives in the upper right part of the page). Once this is done you will see your token in the screen, that's the token we will need in step 12 of this guide.

A bit below that we will see a text field where we need to enter the subdomain name we want (something.duckdns.org) and a place to assign an IP (IPv4), in there we can enter a name for our subdomain and for the IP enter the public IP of our ingess (the one from step 5), then click save/update.

Now we are telling DuckDNS to redirect all the traffic that arrives to that subdomain to the IP we entered, wonderful!

(12) Deploy a DuckDNS cert-manager webhook handler
Now is time to deploy a DuckDNS webhook handler, this is what will add the functionality to cert-manager to manage records in DuckDNS. We can opt to use a helm chart or deploy by cloning the repository where this solution resides, the helm chart didn't work for me so I will be describing the approach using the code in the repo instead.

Let's clone the repository first
git clone https://github.com/ebrianne/cert-manager-webhook-duckdns.git
Now we install it from the cloned repository
cd cert-manager-webhook-duckdns

helm install cert-manager-webhook-duckdns --namespace cert-manager --set duckdns.token='TOKEN_DE_DUCKDNS' --set clusterIssuer.production.create=true --set clusterIssuer.staging.create=true --set clusterIssuer.email='NUESTRO_MAIL' --set logLevel=2 ./deploy/cert-manager-webhook-duckdns
Now we will see we have a new pod in our cert-manager namespace, we can check with the following command
kubectl get pods -n cert-manager
And you will see something like this
NAME                                            READY   STATUS    RESTARTS   AGE
cert-manager-webhook-duckdns-5cdbf66f47-kgt99   1/1     Running   0          56m
(13) ClusterIssuers y detalles de cert-manager
To generate certificates cert-manager has two certificate generators, one is called XXXX-Staging and the other one XXXX-Production. The main difference is that the Production one will provide a certificate that is valid for all web browsers, this is the one we want in our application, but if we are testing and learning we will make mistakes and too many mistakes in the Production one will cause cert-manager to ban us from using the service. To avoid this there is the Staging one which will provide a valid certifcate that our brownsers will take as "valid buuuuuuut" so you will see the padlock and the https but you will see in the certificate description that it is a Staging certificate. With this Staging one we can try and make as many mistakes as we need to fully understand how this works, once done, you simply change the ClusterIssuer to the Production one and you will get a new certifica but for Production and since it was working when you did your tests in Staging this one should not fail.

When we installed the DuckDNS webhook we told the names to use for those ClusterIssuers, here is what I mean:
--set clusterIssuer.production.create=true --set clusterIssuer.staging.create=true
(14) Let's create an ingress resource using the Staging ClusterIssuer
Create a file called staging-ingress.yaml with the following content:
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: echo-https-ingress
  namespace: default
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: cert-manager-webhook-duckdns-staging
    nginx.ingress.kubernetes.io/rewrite-target: /$1
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  tls:
  - hosts:
    - superprueba.duckdns.org
    secretName: superprueba-tls-secret-staging
  rules:
  - host: superprueba.duckdns.org
    http:
      paths:
      - path: /(.*)
        pathType: Prefix
        backend:
          service:
            name: echo
            port:
              number: 80

In this example the subdomain is called superprueba and I am defining to use the clusterissuer cert-manager-webhook-duckdns-staging and to store the certificate in a secret called superprueba-tls-secret, also all https traffic coming from superprueba.duckdns.org needs to be forwarded to service echo on port 80.

The secret name can be anything we want, is not mandatory to contain the name of the subdomain/domain but is a good practice so we can quickly identify what the secret is for.

Another important detail is that the ingress resource has to be defined in the same namespace as the service that will be forwarding traffic to, but the ingress CONTROLLER can be (and is normal to configure in this way) in another different namespace.

Now we apply it with the following command:
kubectil apply -f staging-ingress.yaml
(15) Verify the creation process for our certificate
Now if we run a kubectl get challenge in the same namespace where we deployed the ingress resource we should see something like this:
NAME                                                        STATE     DOMAIN                       AGE
superprueba-tls-secret-staging-6lmxj-668717679-4070204345   pending   superprueba.duckdns.org      4s
This is the process that cert-manager uses to generate the TXT record in DuckDNS and confirm we own the subdomain/domain (basically we provided a valid token), once this process is done and cert-manager confirms we are the owner of the subdomain/domain this challenge is deleted and a certificate and a key are generated and stored in the secret we specified (superprueba-tls-secret-staging) in our case.

If we check the status of our certificate while the challenge is still pending we will see something like this:
NAME                             READY   SECRET                           AGE
superprueba-tls-secret-staging   False    superprueba-tls-secret-staging   7m15s
And once is done and the challenge is deleted we will see something like this:
NAME                             READY   SECRET                           AGE
superprueba-tls-secret-staging   True    superprueba-tls-secret-staging   7m15s
At this point we can verify that we can access our subdomain superprueba-duckdns.org with a browser or using curl.

With curl we would see something like this.
 curl https://superprueba.duckdns.org/
curl: (60) schannel: SEC_E_UNTRUSTED_ROOT (0x80090325) - The certificate chain was issued by an authority that is not trusted.
This is correct, we have a certificate but is not a production ready one, is just one to test the configuration for cert-manager is correct, now we can go and change the clusterissuer from staging to production to obtain a real and valid certificate.

(16) Adjusting our ingress resource to request a production certificate
Now let's crete a new file called production-ingress.yaml with the following content:
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: echo-https-ingress
  namespace: default
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: cert-manager-webhook-duckdns-production
    nginx.ingress.kubernetes.io/rewrite-target: /$1
    nginx.ingress.kubernetes.io/use-regex: "true"
spec:
  tls:
  - hosts:
    - superprueba.duckdns.org
    secretName: superprueba-tls-secret-production
  rules:
  - host: superprueba.duckdns.org
    http:
      paths:
      - path: /(.*)
        pathType: Prefix
        backend:
          service:
            name: echo
            port:
              number: 80
And then let's apply it with:
kubectl apply -f production-ingress.yaml
Once this is done, we can do the same verification steps as before to confirm that the production certificate is issued and stored in our secret and confirm by navigating to our site again
curl https://superprueba.duckdns.org/
Test 123!
(17) Troubleshooting
Ok, this was the article on how to configure and make it work this solution without problems, some times we have a typo, or misconfigure an IP or a wrong name somewhere and is a pain in the neck to know what is wrong if you are just following this tutorial as a first approach to kubernetes.
So here are a few things to check in case something is not working as expected.

a) Check the cert-manager webhook logs, here you will find all the actions that the webhook performs to the duckdns service, if there is a problem with the token you are using, a failure to reach duckdns, etc. here is where you will find that.

b) Check the logs of cert-manager (the core element), those are the pods called cert-manager-XXXX and in here you will find information on what is cert-manager doing, if is requesting a certificate, creating a secret, running a challenge, etc.

c) Verify the logs for the ingress-controller pods, here we can see the requests reaching our cluster, if the requests can't reach our ingress they will not be able to be routed to any service, here we should see the request being ingested.

d) Check the configuration in DuckDNS is pointing to the correct IP as we configured it, this can be done with https://digwebinterface.com/ which is a simple page that you input a domain name and it will return you the IP where it is pointing.
