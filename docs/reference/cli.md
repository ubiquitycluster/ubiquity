# Ubiquity CLI Reference

## Overview
The `ubiquity` CLI is the primary entry point for managing Ubiquity clusters.

## Installation
```
make cli
sudo make install
```

## Commands

### init
Bootstrap Ubiquity configuration. Creates ~/.ubiquity/ with skeleton config.
```
ubiquity init
```

### configure
Interactive configuration wizard for cluster settings.
```
ubiquity configure --domain mycluster.example.com
ubiquity configure --interactive
```

### up
Deploy the full cluster stack (6 phases).
```
ubiquity up --sandbox
ubiquity up --env prod
ubiquity up --skip-security
```

### down
Tear down cluster and cloud resources.
```
ubiquity down
```

### status
Show provisioning state and phase progress.
```
ubiquity status
ubiquity status --plain
```

### logs
Read provisioning logs from state.
```
ubiquity logs
ubiquity logs bootstrap
```

### retry
Retry a failed provisioning phase.
```
ubiquity retry metal
ubiquity retry bootstrap
```

### test
Run the test suite.
```
ubiquity test
ubiquity test --integration
```

### version
Print version information.
```
ubiquity version
ubiquity version --json
```

### info
Show cluster information summary.
```
ubiquity info
```

### health
Check cluster health.
```
ubiquity health
```
