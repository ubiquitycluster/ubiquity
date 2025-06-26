# Post-installation: AWX

AWX provides a great way to integrate with nodes, and orchestrate the configuration management of those nodes.
However, to do this needs a few steps, listed below.

## Get AWX Password
You can retrieve this using the opus container, using the ./scripts/awx-admin-password script:

```console
[root@bootstrap ubiquity]# ./scripts/awx-admin-password
WARNING: AWX admin can do most baremetal node config in the cluster, only use it for just enough initial setup or in emergencies.
5Q95BpvbMwCVRegH7FUGr7iCSy4sdes0gl[
```

## Logon to AWX interface
This is provisioned on your cluster using the `https://awx.<my-cluster-domain>` address. If you use nip.io, it will be something like `https://awx.10-0-0-200.nip.io` for example.

<image logon here>

Once you log on to this page, you will be presented with the dashboard of the current cluster:

<insert AWX dashboard image here>

This looks a little sparse, with only 1 node, and a generic test project. We should probably fix that...

## Setup OAuth integration - Optional

If you wish to allow a group of users to logon and you have this attached to KeyCloak, then you can attach it going through the Generic OIDC settings.

## Setup AWX project

## Configure AWX Token
AWX has the concept of tokens for authentication, so hosts can check-in, execute tasks, etc.

- Go to Resources > Credentials on the left-hand side of the navigation headings. In there you will see a Demo credential and an ansible galaxy credential.

<insert awx image for credentials view>

- Click add

<insert add button pic>

- Once there you are going to give it a name of `Ubiquity` - A description of `Private key - ed25519`, An organization of `Default`, and a Credential Type of `Machine`.
- Then, a section will appear below:

<insert screenshot of key paste area here>

- Simply take your key that opus created (inside your .ssh/ed25519 folder - You did back them up, right?) and paste it into the textbox below.

< show pasted key in a window >

- You can define if you requested a passphrase to specify that on the system.

- Then save - You've now got a machine credential to log onto all of your compute nodes. Don't worry, this key is encrypted:

< show encrypted credential view >

## Setup an Inventory
Next up is to create our default Ubiquity inventory. This is where all our hosts live.

- Click the Resources > Inventories tab

< screenshot of resources-inventories >

- Then we'll click add, and create our Ubiquity inventory. The name will be Ubiquity, Organization is Default, and the only other thing as a placeholder is we'll define the ansible_python_interpreter in the variables section. Hit save.

< screenshot of adding new inventory >

- Note now when you hit save, You'll be presented with more options. We'll need to go into a few of these.

< screenshot of initial newly created inventory >

- Let's create some groups. Click the Groups subheading. Note that there aren't any groups yet... We need to fix that.

< show inventories no groups yet pic >

- Add a new group. Call it compute. At the bottom of this, just like the overall inventory - There is now a section for you to put variables. These are now specific variables for this group. Place any specific variables for your playbooks (an example is inside the ubiq-playbooks repo for defaults) here.

< screenshot of variables >

- Save.

- Repeat this process for: cesgpfs,
We don't explicitly right now get nodes to auto-execute AWX playbooks, but we do enable the
automatic check-in of hosts to an inventory.

## Next-Steps