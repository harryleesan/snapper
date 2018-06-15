## Usage

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
docker run --rm -it -v $(pwd):/go/src/app -v $HOME/.aws:/home/1000/.aws
halosan/snapper-dev:latest bash
```

1. Install dependencies:

    ```bash
    dep ensure
    ```

2. To package this **Go** application using SAM and upload into S3 (Note that
   this will also update the `snapper-serverless.yaml` file):

    ```bash
    sam package --template-file ./template.yaml --output-template-file snapper-serverless.yaml --s3-bucket vatit-lambdas
    ```

3. To deploy the Lambda function through your own CloudFormation stack:

    ```bash
    sam deploy --template-file snapper-serverless.yaml --stack-name snapper
    ```
