# Adracan - Diffs Resources Accross Clusters And Namespaces

Adracan is a CLI tool that helps developers diff Kubernetes resources accross clusters and namespaces. It can be useful when something is broken in production and you want to compare it to the same resource in staging.


## Usage

```console
> adracan context --from stg  --to prd --namespace test deployment foo-bar
```