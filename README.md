#  pbin

- https://gitlab.com/reedrichards/pbin

## Quickstart

if you have [just](https://github.com/casey/just) and [docker](https://docs.docker.com/get-docker/) installed, you can
start the project with `just run`. Otherwise run  

```shell
$ docker run -p 8000:8000 pbin:latest
```

## Setup and Configuration

create an iam user with the following permissions for terraform:

```json
{
   "Version":"2012-10-17",
   "Statement":[
      {
         "Sid":"ListImagesInRepository",
         "Effect":"Allow",
         "Action":[
            "ecr:ListImages"
         ],
         "Resource":"arn:aws:ecs:us-east-1:150301572911:repository/pbin"
      },
      {
         "Sid":"GetAuthorizationToken",
         "Effect":"Allow",
         "Action":[
            "ecr:GetAuthorizationToken"
         ],
         "Resource":"*"
      },
      {
         "Sid":"ManageRepositoryContents",
         "Effect":"Allow",
         "Action":[
                "ecr:*"
         ],
         "Resource":"arn:aws:ecr:us-east-1:150301572911:repository/pbin"
      },
     {
       "Sid": "VisualEditor0",
       "Effect": "Allow",
       "Action": [
         "apprunner:ListConnections",
         "apprunner:ListAutoScalingConfigurations",
         "apprunner:ListServices",
         "iam:*"

       ],
       "Resource": "*"
     },
     {
       "Sid": "VisualEditor1",
       "Effect": "Allow",
       "Action": "apprunner:*",
       "Resource": [
         "arn:aws:apprunner:us-east-1:150301572911:connection/*/*",
         "arn:aws:apprunner:us-east-1:150301572911:autoscalingconfiguration/*/*/*",
         "arn:aws:apprunner:us-east-1:150301572911:service/*/*"
       ]
     }
   ]
}
```

dynamodb iam permissions

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "dynamodb:ListContributorInsights",
                "dynamodb:DescribeReservedCapacityOfferings",
                "dynamodb:ListGlobalTables",
                "dynamodb:ListTables",
                "dynamodb:DescribeReservedCapacity",
                "dynamodb:ListBackups",
                "dynamodb:PurchaseReservedCapacityOfferings",
                "dynamodb:DescribeLimits",
                "dynamodb:ListExports",
                "dynamodb:ListStreams"
            ],
            "Resource": "*"
        },
        {
            "Sid": "VisualEditor1",
            "Effect": "Allow",
            "Action": "dynamodb:*",
            "Resource": "arn:aws:dynamodb:us-east-1:150301572911:table/pbin_*"
        },
        {
            "Sid": "VisualEditor2",
            "Effect": "Allow",
            "Action": "dynamodb:*",
            "Resource": [
                "arn:aws:dynamodb:us-east-1:150301572911:table/pbin_*/stream/*",
                "arn:aws:dynamodb:us-east-1:150301572911:table/pbin_*/export/*",
                "arn:aws:dynamodb::150301572911:global-table/pbin_*",
                "arn:aws:dynamodb:us-east-1:150301572911:table/pbin_*/backup/*",
                "arn:aws:dynamodb:us-east-1:150301572911:table/pbin_*/index/*"
            ]
        }
    ]
}
```

example terraform 

```hcl
resource "aws_iam_user" "pbin" {
  name = "pbin"

  tags = {
    Project  = "pbin"
    Type  = "terraform"
  }
}

data "template_file" "pbin" {
  template = file("./policies/pbin.json")
} 
resource "aws_iam_user_policy" "pbin" {
  name = "pbin"
  user = aws_iam_user.pbin.name

  policy = data.template_file.pbin.rendered
}

```

configure https://gitlab.com/reedrichards/pbin/-/settings/ci_cd
for with access key terraform user 

create access key https://console.aws.amazon.com/iam/home#/users/pbin?section=security_credentials

- ``
- `AWS_SECRET_ACCESS_KEY`

push this repo to the new project on gitlab 

```shell
git init --initial-branch=main
git remote add origin git@gitlab.com:reedrichards/pbin.git
git add .
git commit -m "Initial commit"
git push -u origin main
```

## Deployment

### required environment variables

| key                     | description | required | example                  |
|-------------------------|-------------|----------|--------------------------|
| `AWS_ACCESS_KEY_ID`     |             | yes      |                          |
| `AWS_SECRET_ACCESS_KEY` |             | yes      |                          |
| `AWS_REGION`            |             | yes      | us-east-1                         |
| `PBIN_TABLE_NAME`       |             | yes      | pbin_dev                 |
| `PBIN_URL`              |             | yes      | https://pbin.jjk.is:8000 |

