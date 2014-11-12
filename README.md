## hooks

Hooks is a small application that manages web hooks from github, hub.docker.com, or 
other third party services.


### cli

```bash
NAME:
   hooks - manage github webhooks and events

USAGE:
   hooks [global options] command [command options] [arguments...]

VERSION:
   2

COMMANDS:
   github       handle github webhooks by pushing them onto a queue names hooks-{reponame}
   archive      archive hooks into a rethinkdb for processing
   broadcast    broadcast is a command that accepts jobs off of a queue and sends a hook to third party services
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug              enable debug output
   --config, -c         config file path
   --help, -h           show help
   --version, -v        print the version
```

### license - MIT
