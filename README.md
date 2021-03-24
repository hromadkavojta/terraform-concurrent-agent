##Welcome to terraform concurrent agent
This agent is designed to improve team collaboration with focus on improving GitOps pipeline for terraform on google cloud. This agent works on flexible environment in app engine.

###Setting up

There are 3 parts for installing this agent to your google cloud project

Fist download your preferred terraform binary from https://www.terraform.io/downloads.html and place it the directory `agent` 

### Creating secure connection
- Create secret string that you place into makefile in root directory, into agent-cloudbuild.yaml authorization header and `agent/app.yaml` environment variable

### Github access
- Go to `agent/app.yaml` and fill out environment variables. Then add your superuser id_rsa, id_rsa.pub and known_hosts files to directory `agent`

### Creating GCP project
- Create project on google cloud platform and create account service that has rights to app engine and place your service account in folder `agent`
- Create cloudbuild trigger that will use `agent-cloudbuild.yaml` , this cloudbuild has to be in your future infrastructure repository

For last but not least change url address in root Makefile and in `agent-cloudbuild.yaml` to your future app engine address, URL address should look like `https://agent-dot-your-project-id.ew.r.appspot.com/`

- then you can call `make create` from root directory and infrastructure agent will be deployed and client app will create executable file `tfagent`

`tfagent` has three callings, `./tfagent show`, `./tfagent plan`, `./tfagent apply`

- When commit to master branch is pushed, it automatically creates terraform plan, you can now inspect changes with `./tfagent show` and then apply with `./tfagent apply`

Happy developing!
yes
bachlory work 1