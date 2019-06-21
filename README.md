# azure-pipeline-operator
Kubernetes operator to start Azure Pipeline agents on demand

## AzureAgentPool CRD 
The operator comes with a _AzureAgentPool_ Kuberneted custom resource with the following fields:

* account: Azure Devops organization name
* myproject: Azure Devops project within the organization to watch
* accessToken: Personal Access Token with permissions to manage agents and view build
* agentPool: Name of the agent pool to add agents to

## Setup Dev environment
Install the [Operator Framework](https://github.com/operator-framework/operator-sdk). 

```
#On a Mac
brew install operator-sdk

export GO111MODULE=on
cd $SOURCE_HOME

#Pull dependencies
go mod vendor
```

## Try with MiniShift
```
 oc new-project myproject
 #Install custom resource definition
 oc apply -f deploy/crds/apo_v1alpha1_azureagentpool_crd.yaml
 
 #Create custom resource
 oc apply -f deploy/crds/apo_v1alpha1_azureagentpool_cr.yaml
 
 #Run operator locally
 operator-sdk up local --namespace myproject
```

## Test Azure Pipeline project
First step of the project must be a curl to register the agent that the following tasks will run on. The Operator will actually start this agent on your Kubernetes cluster.
The registrstion is only required so the following steps wait for the agent to come online. Here we need an agent (KUBERNETES_AGENT_TYPE=main) - permanently running - just to execute the curl command:
```
pool:
  name: MyPool
  demands: KUBERNETES_AGENT_TYPE -equals main

steps:
- script: |
   echo Registering agent for $PAT_USERNAME build $BUILD_BUILDID.
   echo Calling $SYSTEM_COLLECTIONURI
   
   #Register agent for this build - agent will start later
   curl -s -X POST -H "Content-Type: application/json" \
   -u anything:$ACCESS_TOKEN \
   -d '
   {
       "maxParallelism": 1,
       "name": "agent-build-'$BUILD_BUILDID'",
       "enabled": true,
       "version": "2.147.1",
       "systemCapabilities": {
           "BUILD_BUILDID": "'$BUILD_BUILDID'"
       }
   }
   ' ${SYSTEM_COLLECTIONURI}_apis/distributedtask/pools/9/agents?api-version=5.1-preview.1
  displayName: 'Command Line Script'

```

The agent created by the operator will have the matching name and BUILD_BUILDID capability. So a task running on that agent should be defined like this:
```
dependsOn: Job_3
pool:
  name: MyPool
  demands: BUILD_BUILDID -equals $(Build.BuildId)

  timeoutInMinutes: 1

steps:
- script: |
   echo HELLO IS RUNNING
   
  displayName: 'Command Line Script'
```
