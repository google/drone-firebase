# drone-firebase

Drone plugin to deploy a project to Firebase. For the usage information
and a listing of the available options please take a look at
[the docs](DOCS.md).

*This is not an official Google product.*

If you want to contribute, see the [CONTRIBUTING page](CONTRIBUTING.md).

## Binary

Build the binary using `make`:

```
make deps build
```

### Example

```sh
./drone-firebase -- "{
    \"system\": {
        \"link\": \"http://drone.mycompany.com\"
    },
    \"repo\": {
        \"owner\": \"octocat\",
        \"name\": \"hello-world\",
        \"full_name\": \"octocat/hello-world\",
        \"link_url\": \"https://github.com/octocat/hello-world\",
        \"clone_url\": \"https://github.com/octocat/hello-world.git\"
    },
    \"build\": {
        \"number\": 1,
        \"event\": \"push\",
        \"branch\": \"master\",
        \"commit\": \"436b7a6e2abaddfd35740527353e78a227ddcb2c\",
        \"ref\": \"refs/heads/master\",
        \"author\": \"octocat\",
        \"author_email\": \"octocat@github.com\"
    },
    \"workspace\": {
        \"root\": \"/drone/src\",
        \"path\": \"/drone/src/github.com/octocat/hello-world\",
        \"keys\": {
            \"private\": \"-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQC...\"
        }
    },
    \"vargs\": {
        \"token\": \"thetoken\",
        \"project_id\": \"my-cool-project\"
    }
}"
```

## Docker

Build the container using `make`:

```
make deps docker
```

### Example

Run the docker container from the directory of your Firebase project:

```sh
docker run -i -v $(pwd):/drone/src/github.com/octocat/hello-world google/drone-firebase -- "{
    \"system\": {
        \"link\": \"http://drone.mycompany.com\"
    },
    \"repo\": {
        \"owner\": \"octocat\",
        \"name\": \"hello-world\",
        \"full_name\": \"octocat/hello-world\",
        \"link_url\": \"https://github.com/octocat/hello-world\",
        \"clone_url\": \"https://github.com/octocat/hello-world.git\"
    },
    \"build\": {
        \"number\": 1,
        \"event\": \"push\",
        \"branch\": \"master\",
        \"commit\": \"436b7a6e2abaddfd35740527353e78a227ddcb2c\",
        \"ref\": \"refs/heads/master\",
        \"author\": \"octocat\",
        \"author_email\": \"octocat@github.com\"
    },
    \"workspace\": {
        \"root\": \"/drone/src\",
        \"path\": \"/drone/src/github.com/octocat/hello-world\",
        \"keys\": {
            \"private\": \"-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQC...\"
        }
    },
    \"vargs\": {
        \"token\": \"thetoken\",
        \"project_id\": \"my-cool-project\"
    }
}"
```
