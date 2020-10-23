# kube-database-creator

Automates the creation of application databases in a microservice environments.

One of the key aspects of a microservice architecture is that each server should have its own database isolated from all the other services. So when adding a new microservice to the system there is usually a task to create a new database and database-user for this service. 

`kube-database-creator` automates this task, in the regard that a service just has declare its need for a database via a kubernetes ConfigMap, the creation of the database itself together with the required credentials.

Note: The tool just *creates* databases. It is not supposed to automatically clean it up afterwards, i.e. it will not delete or try to delete anything. (Just in case you are worring about losing valuable data.)

## Current state

This is a very basic implementation only supporting postgres so far.

Things that need to be added:
* A secrets store backend for HashiCorp vault (so that the application not even sees the admin-user-credentials)
* Support for some other common database: MySql, MariaDB, you name it.
* Support for monitoring/altering in case something went wrong during creation.

## Example

An example configuration can be found in `example/example.yaml`. Which can be used like this:

```
kubectl apply -f example/example.yaml
```

Which will create a namespace `creator-example` containing a postgres database and a `kube-database-creator`.

Initially you can connect to the postgres instance via port-forwarding:
```
kubectl -n creator-example port-forward service/postgres 5432
```

and then (in another console or with a postgres toll of your choice):
```
psql -h localhost -U postgres -W postgres
```
(The password is `verysecretmasterpassword` btw.)

Initially it should just contains the basic postgres tables.

Now imaging a new microservice `demo-app` in need of its own datbase. In this case one just has to do:
```
kubectl apply -f example/demo-app.yaml
```

If you now look into the postgres again you will see that a database `demo_app_db` has been created, together with a kubernetes secret to access it. The database password for the application is randomly generated and only available via that kubernetes secret.

To check if you can connect into the `demo-app` with:
```
kubectl -n creator-example exec -ti demo-app -- bash
```

and do a:
```
PGPASSWORD=$DEMO_APP_DB_PASSWORD psql -h postgres.creator-example.svc.cluster.local -U $DEMO_APP_DB_USER $DEMO_APP_DB_NAME
```
inside.

To clean this mess up just do a
```
kubectl delete -f example/example.yaml
```
