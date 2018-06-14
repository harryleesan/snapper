## Usage

### Pre-requisites
- **Go** (Of course)
  - [dep](https://github.com/golang/dep) (Dependency management)
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-welcome.html<Paste>) (Just need your AWS creds)
- [SAM CLI](https://github.com/awslabs/aws-sam-cli) (For packaging this
  lambda application).

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
