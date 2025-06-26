<!-- omit in toc -->
# Deploy Private Git Repository using Gitea

You can deploy your own private Git repository using [Gitea](https://gitea.io/en-us/) to use AWX with the playbooks on SCM.

However, it just so happens... We've already done that for you by-default in Ubiquity! Aren't we good to you!

<!-- omit in toc -->
## Table of Contents

- [Configure AWX to use Git Repository with Self-Signed Certificate](#configure-awx-to-use-git-repository-with-self-signed-certificate)

## Configure AWX to use Git Repository with Self-Signed Certificate

1. Add Credentials for SCM
2. Allow Self-Signed Certificate such as this Gitea
   - Open `Settings` > `Jobs settings` in AWX
   - Press `Edit` and scroll down to `Extra Environment Variables`, then add `"GIT_SSL_NO_VERIFY": "True"` in `{}`

     ```json
     {
       "GIT_SSL_NO_VERIFY": "True"
     }
     ```

   - Press `Save`
