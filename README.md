## Introduction

**Snapper-go** is a simple snapshot helper that takes snapshots of EBS volumes
of tagged instances on a cron schedule. This application also removes snapshots
that are over a specific age.

### How it works
This **Go** application is deployed as a _AWS Lambda_ function which when called
checks **all** _AWS EC2_ instances for the _Snapper_ tag. It then creates (or deletes
_AWS EBS_ snapshots) depending on the input. This _Lambda_ function is called
via _AWS CloudWatch_ scheduled events.

For ease of use, it is deployed as a _AWS CloudFormation_ stack. See usage
below.

## Usage

To enable this functionality, tag the EC2 instances with the Key _Snapper_ and
the Value with the number of days that you want the snapshots to exist.

```json
{
  "Key": "Snapper",
  "Value": "7"
}
```

### Pre-requisites
- **Go** (Of course)
  - [dep](https://github.com/golang/dep) (Dependency management)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-welcome.html<Paste>) (Just need your AWS creds)
- [SAM CLI](https://github.com/awslabs/aws-sam-cli) (For packaging this
  lambda application).

This application can be developed in **Docker** (See _Dockerfile_) without having
the above pre-requisites installed (All the tools are packaged in the docker
image).

```bash
docker build -t halosan/snapper-dev:latest .
docker run --rm -it -v $(pwd):/go/src/app -v $HOME/.aws:/home/1000/.aws halosan/snapper-dev:latest bash
```

(It is recommended to develop in the provided docker container, this ensures that
you don't have version issues.)

1. Install dependencies:

    ```bash
    dep ensure
    ```
2. Build the binary for the **Lambda function**:

    ```bash
    go build main.go
    ```

3. To package this **Go** application using **SAM** and upload into **S3** (Note that
   this will also update the `snapper-serverless.yaml` file):

    ```bash
    sam package --template-file ./template.yaml --output-template-file snapper-serverless.yaml \
    --s3-bucket vatit-lambdas --s3-prefix snapper-go
    ```

4. To deploy the Lambda function through your own CloudFormation stack:

    ```bash
    sam deploy --template-file snapper-serverless.yaml --stack-name snapper
    ```
